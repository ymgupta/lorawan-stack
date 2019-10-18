// Copyright © 2019 The Things Industries B.V.

package stripe

import (
	"context"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/client"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

const (
	managedTenantAttribute        = "managed-by"
	stripeManagerAttributeValue   = "stripe"
	stripeCustomerIDAttribute     = "stripe-customer-id"
	stripePlanIDAttribute         = "stripe-plan-id"
	stripeSubscriptionIDAttribute = "stripe-subscription-id"
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
	_, err = client.Create(ctx, &ttipb.CreateTenantRequest{
		Tenant: ttipb.Tenant{
			TenantIdentifiers: ttipb.TenantIdentifiers{
				TenantID: convertCustomerNameToTenantID(customer.Name),
			},
			Name:        customer.Name,
			Description: customer.Description,
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
		},
	}, s.tenantAuth)
	if err != nil {
		return err
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
				TenantID: convertCustomerNameToTenantID(customer.Name),
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

func (s *Stripe) deleteTenant(ctx context.Context, sub *stripe.Subscription) error {
	client, err := s.getTenantRegistry(ctx)
	if err != nil {
		return err
	}
	customer, err := s.getCustomer(sub.Customer.ID, nil)
	if err != nil {
		return err
	}
	_, err = client.Delete(ctx, &ttipb.TenantIdentifiers{
		TenantID: convertCustomerNameToTenantID(customer.Name),
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

func (s *Stripe) getAPIClient() (*client.API, error) {
	if s.apiClient != nil {
		return s.apiClient, nil
	}
	backends := stripe.NewBackends(nil)
	backends.API = stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
		LeveledLogger: log.FromContext(s.ctx),
		LogLevel:      s.config.LogLevel,
	})
	return client.New(s.config.APIKey, backends), nil
}

func (s *Stripe) getCustomer(id string, params *stripe.CustomerParams) (*stripe.Customer, error) {
	client, err := s.getAPIClient()
	if err != nil {
		return nil, err
	}
	return client.Customers.Get(id, params)
}

func convertCustomerNameToTenantID(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), " ", "")
}
