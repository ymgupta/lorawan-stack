// Copyright © 2019 The Things Industries B.V.

package stripe

import (
	"context"
	"strconv"

	"github.com/gogo/protobuf/types"
	"github.com/stripe/stripe-go"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

const (
	managedTenantAttribute            = "managed-by"
	stripeManagerAttributeValue       = "stripe"
	stripeCustomerIDAttribute         = "stripe-customer-id"
	stripePlanIDAttribute             = "stripe-plan-id"
	stripeSubscriptionIDAttribute     = "stripe-subscription-id"
	stripeSubscriptionItemIDAttribute = "stripe-subscription-item-id"

	tenantIDStripeAttribute         = "tenant-id"
	maxApplicationsStripeAttribute  = "max-applications"
	maxClientsStripeAttribute       = "max-clients"
	maxEndDevicesStripeAttribute    = "max-end-devices"
	maxGatewaysStripeAttribute      = "max-gateways"
	maxOrganizationsStripeAttribute = "max-organizations"
	maxUsersStripeAttribute         = "max-users"
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
			attribute: maxApplicationsStripeAttribute,
			value:     &tnt.MaxApplications,
		},
		{
			attribute: maxClientsStripeAttribute,
			value:     &tnt.MaxClients,
		},
		{
			attribute: maxEndDevicesStripeAttribute,
			value:     &tnt.MaxEndDevices,
		},
		{
			attribute: maxGatewaysStripeAttribute,
			value:     &tnt.MaxGateways,
		},
		{
			attribute: maxOrganizationsStripeAttribute,
			value:     &tnt.MaxOrganizations,
		},
		{
			attribute: maxUsersStripeAttribute,
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

	ids, err := generateTenantID(sub)
	if err != nil {
		return err
	}
	tnt := ttipb.Tenant{
		TenantIdentifiers: *ids,
		Name:              customer.Name,
		Description:       customer.Description,
		Attributes: map[string]string{
			managedTenantAttribute:            stripeManagerAttributeValue,
			stripeCustomerIDAttribute:         customer.ID,
			stripePlanIDAttribute:             sub.Plan.ID,
			stripeSubscriptionIDAttribute:     sub.ID,
			stripeSubscriptionItemIDAttribute: sub.Items.Data[0].ID,
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
	_, err = client.Get(ctx, &ttipb.GetTenantRequest{TenantIdentifiers: *ids, FieldMask: tntFieldMask}, s.tenantAuth)
	if err != nil && !errors.IsNotFound(err) {
		return err
	} else if err == nil {
		tntExists = true
	}
	if tntExists {
		_, err := client.Update(ctx, &ttipb.UpdateTenantRequest{
			Tenant:    tnt,
			FieldMask: tntFieldMask,
		}, s.tenantAuth)
		if err != nil {
			return err
		}
	} else {
		initialUser, err := s.generateInitialUser(ctx, sub, customer)
		if err != nil {
			return err
		}
		_, err = client.Create(ctx, &ttipb.CreateTenantRequest{
			Tenant:      tnt,
			InitialUser: initialUser,
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
	ids, err := generateTenantID(sub)
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

func generateTenantID(sub *stripe.Subscription) (*ttipb.TenantIdentifiers, error) {
	if tenantID, ok := sub.Metadata[tenantIDStripeAttribute]; ok {
		return &ttipb.TenantIdentifiers{
			TenantID: tenantID,
		}, nil
	}
	return nil, errNoTenantID
}
