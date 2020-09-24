// Copyright © 2019 The Things Industries B.V.

package stripe

import (
	"context"
	"strconv"

	"github.com/gogo/protobuf/types"
	"github.com/stripe/stripe-go"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

const (
	adminEmailAttribute        = "admin-email"
	adminFullNameAttribute     = "admin-full-name"
	adminPasswordAttribute     = "admin-password"
	adminUserIDAttribute       = "admin-user"
	companyAttribute           = "company"
	maxApplicationsAttribute   = "max-applications"
	maxClientsAttribute        = "max-clients"
	maxEndDevicesAttribute     = "max-end-devices"
	maxGatewaysAttribute       = "max-gateways"
	maxOrganizationsAttribute  = "max-organizations"
	maxUsersAttribute          = "max-users"
	nameAttribute              = "name"
	tenantDescriptionAttribute = "tenant-description"
	tenantIDAttribute          = "tenant-id"
	tenantNameAttribute        = "tenant-name"
)

var (
	errTenantNotManaged   = errors.DefineFailedPrecondition("tenant_not_managed", "tenant is not managed by Stripe")
	errCustomerIDMismatch = errors.DefineFailedPrecondition("customer_id_mismatch", "tenant is owned by another customer")
	errNoManagedPlan      = errors.DefineFailedPrecondition("no_managed_plan", "no managed plan in the subscription")
)

func (s *Stripe) addTenantLimits(tnt *ttipb.Tenant, sub *stripe.Subscription) error {
	subscriptionItem := s.findSubscriptionItem(sub.Items)
	if subscriptionItem == nil {
		return errNoManagedPlan
	}
	plan := subscriptionItem.Plan

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

	subscriptionItem := s.findSubscriptionItem(sub.Items)
	if subscriptionItem == nil {
		return errNoManagedPlan
	}

	tnt, err := s.toTenant(sub, subscriptionItem, customer)
	if err != nil {
		return err
	}
	ctx = log.NewContextWithField(ctx, "tenant_id", tnt.TenantID)

	err = s.addTenantLimits(tnt, sub)
	if err != nil {
		return err
	}

	tntFieldMask := types.FieldMask{
		Paths: []string{
			"billing",
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
	existingTnt, err := client.Get(ctx, &ttipb.GetTenantRequest{TenantIdentifiers: tnt.TenantIdentifiers, FieldMask: tntFieldMask}, s.tenantAuth)
	if err != nil && !errors.IsNotFound(err) {
		return err
	} else if err == nil {
		tntExists = true
	}

	if tntExists {
		billing := existingTnt.Billing.GetStripe()
		if billing == nil {
			return errTenantNotManaged.New()
		}
		if billing.CustomerID != customer.ID {
			return errCustomerIDMismatch.New()
		}

		_, err := client.Update(ctx, &ttipb.UpdateTenantRequest{
			Tenant:    *tnt,
			FieldMask: tntFieldMask,
		}, s.tenantAuth)
		if err != nil {
			return err
		}
	} else {
		password, _ := sub.Metadata[adminPasswordAttribute]
		name, _ := sub.Metadata[adminFullNameAttribute]
		userID, ok := sub.Metadata[adminUserIDAttribute]
		if !ok {
			userID = tnt.TenantID
		}
		email, ok := sub.Metadata[adminEmailAttribute]
		if !ok {
			email = customer.Email
		}

		_, err = client.Create(ctx, &ttipb.CreateTenantRequest{
			Tenant: *tnt,
			InitialUser: &ttnpb.User{
				UserIdentifiers: ttnpb.UserIdentifiers{
					UserID: userID,
				},
				Name:                name,
				PrimaryEmailAddress: email,
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

func (s *Stripe) updateTenantState(ctx context.Context, sub *stripe.Subscription, state ttnpb.State) error {
	client, err := s.getTenantRegistry(ctx)
	if err != nil {
		return err
	}

	ids, err := toTenantIDs(sub)
	if err != nil {
		return err
	}
	ctx = log.NewContextWithField(ctx, "tenant_id", ids.TenantID)
	tnt, err := client.Get(ctx, &ttipb.GetTenantRequest{
		TenantIdentifiers: *ids,
		FieldMask: types.FieldMask{
			Paths: []string{
				"billing",
				"state",
			},
		},
	}, s.tenantAuth)
	if err != nil {
		return err
	}

	billing := tnt.Billing.GetStripe()
	if billing == nil {
		return errTenantNotManaged.New()
	}
	if billing.CustomerID != sub.Customer.ID {
		return errCustomerIDMismatch.New()
	}

	if tnt.State == state {
		// If the tenant is already in that state, do not attempt an update.
		return nil
	}
	_, err = client.Update(ctx, &ttipb.UpdateTenantRequest{
		Tenant: ttipb.Tenant{
			TenantIdentifiers: *ids,
			State:             state,
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

func (s *Stripe) toTenant(sub *stripe.Subscription, subscriptionItem *stripe.SubscriptionItem, customer *stripe.Customer) (*ttipb.Tenant, error) {
	ids, err := toTenantIDs(sub)
	if err != nil {
		return nil, err
	}

	var tenantName string
	if name, ok := sub.Metadata[tenantNameAttribute]; ok {
		tenantName = name
	} else if name, ok := sub.Metadata[nameAttribute]; ok {
		tenantName = name
	} else if company, ok := sub.Metadata[companyAttribute]; ok {
		tenantName = company
	} else {
		tenantName = customer.Name
	}

	var tenantDescription string
	if description, ok := sub.Metadata[tenantDescriptionAttribute]; ok {
		tenantDescription = description
	} else {
		tenantDescription = customer.Description
	}

	contactInfo := []*ttnpb.ContactInfo{
		{
			ContactType:   ttnpb.CONTACT_TYPE_BILLING,
			ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
			Value:         customer.Email,
			Public:        false,
		},
	}
	if adminEmail, ok := sub.Metadata[adminEmailAttribute]; ok {
		contactInfo = append(contactInfo, &ttnpb.ContactInfo{
			ContactType:   ttnpb.CONTACT_TYPE_TECHNICAL,
			ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
			Value:         adminEmail,
			Public:        false,
		})
	}

	return &ttipb.Tenant{
		TenantIdentifiers: *ids,
		Name:              tenantName,
		Description:       tenantDescription,
		ContactInfo:       contactInfo,
		State:             ttnpb.STATE_APPROVED,
		Billing: &ttipb.Billing{
			Provider: &ttipb.Billing_Stripe_{
				Stripe: &ttipb.Billing_Stripe{
					CustomerID:         customer.ID,
					PlanID:             subscriptionItem.Plan.ID,
					SubscriptionID:     sub.ID,
					SubscriptionItemID: subscriptionItem.ID,
				},
			},
		},
	}, nil
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
