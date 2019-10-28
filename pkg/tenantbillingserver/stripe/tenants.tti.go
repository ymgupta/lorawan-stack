// Copyright © 2019 The Things Industries B.V.

package stripe

import (
	"context"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/stripe/stripe-go"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

const (
	managedTenantAttribute        = "managed-by"
	stripeManagerAttributeValue   = "stripe"
	stripeCustomerIDAttribute     = "stripe-customer-id"
	stripePlanIDAttribute         = "stripe-plan-id"
	stripeSubscriptionIDAttribute = "stripe-subscription-id"
	tenantIDStripeAttribute       = "tenant-id"
)

func (s *Stripe) createTenant(ctx context.Context, sub *stripe.Subscription) error {
	client, err := s.getTenantRegistry(ctx)
	if err != nil {
		return err
	}
	customer, err := s.getCustomer(sub.Customer.ID, nil)
	if err != nil {
		return err
	}

	ids := ttipb.TenantIdentifiers{
		TenantID: generateTenantID(customer, sub),
	}
	tnt := ttipb.Tenant{
		TenantIdentifiers: ids,
		Name:              customer.Name,
		Description:       customer.Description,
		Attributes: map[string]string{
			managedTenantAttribute:        stripeManagerAttributeValue,
			stripeCustomerIDAttribute:     customer.ID,
			stripePlanIDAttribute:         sub.Plan.ID,
			stripeSubscriptionIDAttribute: sub.ID,
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
	tntFieldMask := types.FieldMask{
		Paths: []string{
			"attributes",
			"contact_info",
			"description",
			"name",
			"state",
		},
	}

	var tntExists bool
	_, err = client.Get(ctx, &ttipb.GetTenantRequest{TenantIdentifiers: ids, FieldMask: tntFieldMask}, s.tenantAuth)
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
		_, err = client.Create(ctx, &ttipb.CreateTenantRequest{
			Tenant: tnt,
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
	customer, err := s.getCustomer(sub.Customer.ID, nil)
	if err != nil {
		return err
	}
	_, err = client.Update(ctx, &ttipb.UpdateTenantRequest{
		Tenant: ttipb.Tenant{
			TenantIdentifiers: ttipb.TenantIdentifiers{
				TenantID: generateTenantID(customer, sub),
			},
			State: ttnpb.STATE_SUSPENDED,
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

func generateTenantID(customer *stripe.Customer, sub *stripe.Subscription) string {
	if tenantID, ok := sub.Metadata[tenantIDStripeAttribute]; ok {
		return tenantID
	}
	return strings.ReplaceAll(strings.ToLower(customer.Name), " ", "")
}
