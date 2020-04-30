// Copyright © 2019 The Things Industries B.V.

package stripe

import (
	"context"
	"strconv"

	"github.com/gogo/protobuf/types"
	"github.com/stripe/stripe-go"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/tenantbillingserver/backend"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

const (
	managerAttributeValue       = "stripe"
	customerIDAttribute         = "stripe-customer-id"
	planIDAttribute             = "stripe-plan-id"
	subscriptionIDAttribute     = "stripe-subscription-id"
	subscriptionItemIDAttribute = "stripe-subscription-item-id"

	tenantIDAttribute         = "tenant-id"
	maxApplicationsAttribute  = "max-applications"
	maxClientsAttribute       = "max-clients"
	maxEndDevicesAttribute    = "max-end-devices"
	maxGatewaysAttribute      = "max-gateways"
	maxOrganizationsAttribute = "max-organizations"
	maxUsersAttribute         = "max-users"
	adminUserIDAttribute      = "admin-user"
	adminPasswordAttribute    = "admin-password"
)

var (
	errTenantNotManaged   = errors.DefineFailedPrecondition("tenant_not_managed", "tenant is not managed by Stripe")
	errCustomerIDMismatch = errors.DefineFailedPrecondition("customer_id_mismatch", "tenant is owned by another customer")
)

func (s *Stripe) addTenantLimits(tnt *ttipb.Tenant, sub *stripe.Subscription) error {
	plan, err := s.getPlan(sub.Plan.ID, nil)
	if err != nil {
		return err
	}

	for _, field := range []struct {
		attribute string
		value     **types.UInt64Value
	}{
		{
			attribute: maxApplicationsAttribute,
			value:     &tnt.MaxApplications,
		},
		{
			attribute: maxClientsAttribute,
			value:     &tnt.MaxClients,
		},
		{
			attribute: maxEndDevicesAttribute,
			value:     &tnt.MaxEndDevices,
		},
		{
			attribute: maxGatewaysAttribute,
			value:     &tnt.MaxGateways,
		},
		{
			attribute: maxOrganizationsAttribute,
			value:     &tnt.MaxOrganizations,
		},
		{
			attribute: maxUsersAttribute,
			value:     &tnt.MaxUsers,
		},
	} {
		if v, ok := plan.Metadata[field.attribute]; ok {
			limit, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return err
			}
			*field.value = &types.UInt64Value{Value: limit}
		}
	}

	return nil
}

func (s *Stripe) createTenant(ctx context.Context, sub *stripe.Subscription) error {
	client, err := s.getTenantRegistry(ctx)
	if err != nil {
		return err
	}
	customer, err := s.getCustomer(sub.Customer.ID, nil)
	if err != nil {
		return err
	}

	ids, err := toTenantIDs(sub)
	if err != nil {
		return err
	}
	tnt := ttipb.Tenant{
		TenantIdentifiers: *ids,
		Name:              customer.Name,
		Description:       customer.Description,
		Attributes: map[string]string{
			backend.ManagedByTenantAttribute: managerAttributeValue,
			customerIDAttribute:              customer.ID,
			planIDAttribute:                  sub.Plan.ID,
			subscriptionIDAttribute:          sub.ID,
			subscriptionItemIDAttribute:      sub.Items.Data[0].ID,
		},
		ContactInfo: []*ttnpb.ContactInfo{
			{
				ContactType:   ttnpb.CONTACT_TYPE_BILLING,
				ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
				Value:         customer.Email,
				Public:        false,
			},
		},
		State: ttnpb.STATE_APPROVED,
	}

	err = s.addTenantLimits(&tnt, sub)
	if err != nil {
		return err
	}

	tntFieldMask := types.FieldMask{
		Paths: []string{
			"attributes",
			"contact_info",
			"description",
			"max_applications",
			"max_clients",
			"max_end_devices",
			"max_gateways",
			"max_organizations",
			"max_users",
			"name",
			"state",
		},
	}

	var tntExists bool
	existingTnt, err := client.Get(ctx, &ttipb.GetTenantRequest{TenantIdentifiers: *ids, FieldMask: tntFieldMask}, s.tenantAuth)
	if err != nil && !errors.IsNotFound(err) {
		return err
	} else if err == nil {
		tntExists = true
	}

	if tntExists {
		manager, ok := existingTnt.Attributes[backend.ManagedByTenantAttribute]
		if !ok || manager != managerAttributeValue {
			return errTenantNotManaged.New()
		}
		customerID, ok := existingTnt.Attributes[customerIDAttribute]
		if !ok {
			return errNoTenantAttribute.WithAttributes("attribute", customerIDAttribute)
		}
		if customerID != customer.ID {
			return errCustomerIDMismatch.New()
		}

		_, err := client.Update(ctx, &ttipb.UpdateTenantRequest{
			Tenant:    tnt,
			FieldMask: tntFieldMask,
		}, s.tenantAuth)
		if err != nil {
			return err
		}
	} else {
		password, _ := sub.Metadata[adminPasswordAttribute]
		userID, ok := sub.Metadata[adminUserIDAttribute]
		if !ok {
			userID = ids.TenantID
		}

		_, err = client.Create(ctx, &ttipb.CreateTenantRequest{
			Tenant: tnt,
			InitialUser: &ttnpb.User{
				UserIdentifiers: ttnpb.UserIdentifiers{
					UserID: userID,
				},
				PrimaryEmailAddress: customer.Email,
				State:               ttnpb.STATE_APPROVED,
				Password:            password,
				Admin:               true,
			},
		}, s.tenantAuth)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Stripe) suspendTenant(ctx context.Context, sub *stripe.Subscription) error {
	client, err := s.getTenantRegistry(ctx)
	if err != nil {
		return err
	}
	ids, err := toTenantIDs(sub)
	if err != nil {
		return err
	}
	_, err = client.Update(ctx, &ttipb.UpdateTenantRequest{
		Tenant: ttipb.Tenant{
			TenantIdentifiers: *ids,
			State:             ttnpb.STATE_SUSPENDED,
		},
		FieldMask: types.FieldMask{
			Paths: []string{
				"state",
			},
		},
	}, s.tenantAuth)
	if err != nil {
		return err
	}
	return nil
}

func (s *Stripe) getTenantRegistry(ctx context.Context) (ttipb.TenantRegistryClient, error) {
	if s.tenantsClient != nil {
		return s.tenantsClient, nil
	}
	cc, err := s.component.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		return nil, err
	}
	return ttipb.NewTenantRegistryClient(cc), nil
}

var errNoTenantID = errors.DefineInvalidArgument("no_tenant_id", "no tenant ID set")

func toTenantIDs(sub *stripe.Subscription) (*ttipb.TenantIdentifiers, error) {
	if tenantID, ok := sub.Metadata[tenantIDAttribute]; ok {
		return &ttipb.TenantIdentifiers{
			TenantID: tenantID,
		}, nil
	}
	return nil, errNoTenantID.New()
}
