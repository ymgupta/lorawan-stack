// Copyright © 2019 The Things Industries B.V.

package stripe

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strings"

	echo "github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/web"
)

// Config is the configuration for the Stripe payment backend.
type Config struct {
	Enabled           bool     `name:"enabled" description:"Enable the backend"`
	APIKey            string   `name:"api-key" description:"The API key used to report the metrics"`
	EndpointSecretKey string   `name:"endpoint-secret-key" description:"The endopoint secret used to verify webhook signatures"`
	PlanIDs           []string `name:"plan-ids" description:"The IDs of the subscription plans to be handled"`
}

var (
	errNoAPIKey    = errors.DefineInvalidArgument("no_api_key", "no API key provided")
	errNoSecretKey = errors.DefineInvalidArgument("no_endpoint_secret_key", "no endpoint secret key provided")
	errNoPlanIDs   = errors.DefineInvalidArgument("no_plan_ids", "no plan ids provided")
)

// New returns a new Stripe backend using the config.
func (c Config) New(component *component.Component) (*Stripe, error) {
	if !c.Enabled {
		return nil, nil
	}
	if len(c.APIKey) == 0 {
		return nil, errNoAPIKey
	}
	if len(c.EndpointSecretKey) == 0 {
		return nil, errNoSecretKey
	}
	if len(c.PlanIDs) == 0 {
		return nil, errNoPlanIDs
	}
	return New(component, &c)
}

// Stripe is the payment and tenant configuration backend.
type Stripe struct {
	component *component.Component
	config    *Config
}

// New returns a new Stripe backend.
func New(component *component.Component, config *Config) (*Stripe, error) {
	return &Stripe{component, config}, nil
}

// Report updates the Stripe subscription of the tenant if the tenant is managed by Stripe.
func (s *Stripe) Report(ctx context.Context, tenant *ttipb.Tenant, totals *ttipb.TenantRegistryTotals) error {
	return nil
}

const (
	customerSubscriptionCreated = "customer.subscription.created"
	customerSubscriptionUpdated = "customer.subscription.updated"
	customerSubscriptionDeleted = "customer.subscription.deleted"

	stripeSignatureHeader = "Stripe-Signature"

	managedTenantAttribute        = "managed-by"
	stripeManagerAttributeValue   = "stripe"
	stripeCustomerIDAttribute     = "stripe-customer-id"
	stripeProductIDAttribute      = "stripe-product-id"
	stripePlanIDAttribute         = "stripe-plan-id"
	stripeSubscriptionIDAttribute = "stripe-subscription-id"
)

func (s *Stripe) createTenant(ctx context.Context, sub *stripe.Subscription) error {
	cc, err := s.component.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		return err
	}
	client := ttipb.NewTenantRegistryClient(cc)
	_, err = client.Create(ctx, &ttipb.CreateTenantRequest{
		Tenant: ttipb.Tenant{
			TenantIdentifiers: ttipb.TenantIdentifiers{
				TenantID: strings.ToLower(sub.Customer.Name),
			},
			Name:        sub.Customer.Name,
			Description: sub.Customer.Description,
			Attributes: map[string]string{
				managedTenantAttribute:        stripeManagerAttributeValue,
				stripeCustomerIDAttribute:     sub.Customer.ID,
				stripeProductIDAttribute:      sub.Plan.Product.ID,
				stripePlanIDAttribute:         sub.Plan.ID,
				stripeSubscriptionIDAttribute: sub.ID,
			},
			ContactInfo: []*ttnpb.ContactInfo{
				{
					ContactType:   ttnpb.CONTACT_TYPE_BILLING,
					ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
					Value:         sub.Customer.Email,
					Public:        false,
				},
			},
			State: ttnpb.STATE_APPROVED,
		},
	}, s.component.WithClusterAuth())
	if err != nil {
		return err
	}
	return nil
}

func (s *Stripe) suspendTenant(ctx context.Context, sub *stripe.Subscription) error {
	cc, err := s.component.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		return err
	}
	client := ttipb.NewTenantRegistryClient(cc)
	_, err = client.Update(ctx, &ttipb.UpdateTenantRequest{
		Tenant: ttipb.Tenant{
			TenantIdentifiers: ttipb.TenantIdentifiers{
				TenantID: strings.ToLower(sub.Customer.Name),
			},
			State: ttnpb.STATE_SUSPENDED,
		},
	}, s.component.WithClusterAuth())
	if err != nil {
		return err
	}
	return nil
}

var errInvalidEventType = errors.DefineInvalidArgument("invalid_event_type", "invalid event type `{type}` provided")

func (s *Stripe) handleSubscriptions(c echo.Context) error {
	ctx := s.component.FillContext(c.Request().Context())
	logger := log.FromContext(ctx)

	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}

	event, err := webhook.ConstructEvent(body, c.Request().Header.Get(stripeSignatureHeader), s.config.EndpointSecretKey)
	if err != nil {
		logger.WithError(err).Warn("Webhook signature validation failed")
		return err
	}

	switch event.Type {
	case customerSubscriptionCreated:
	case customerSubscriptionUpdated:
	case customerSubscriptionDeleted:
		break
	default:
		logger.WithFields(log.Fields(
			"event_id", event.ID,
			"event_type", event.Type,
		)).Warn("Unexpected event received")
		return errInvalidEventType.WithAttributes("type", event.Type)
	}

	sub := &stripe.Subscription{}
	if err = json.Unmarshal(event.Data.Raw, sub); err != nil {
		return err
	}

	switch event.Type {
	case customerSubscriptionCreated:
	case customerSubscriptionUpdated:
	case customerSubscriptionDeleted:
		switch sub.Status {
		case stripe.SubscriptionStatusActive:
		case stripe.SubscriptionStatusTrialing:
			return s.createTenant(ctx, sub)
		case stripe.SubscriptionStatusIncomplete:
		case stripe.SubscriptionStatusIncompleteExpired:
			return nil
		case stripe.SubscriptionStatusCanceled:
		case stripe.SubscriptionStatusPastDue:
		case stripe.SubscriptionStatusUnpaid:
			return s.suspendTenant(ctx, sub)
		default:
			panic("unreachable")
		}
	default:
		panic("unreachable")
	}
	return nil
}

// RegisterRoutes implements web.Registerer.
func (s *Stripe) RegisterRoutes(srv *web.Server) {
	group := srv.Group(ttnpb.HTTPAPIPrefix + "/tbs/stripe")
	group.Any("/subscriptions", s.handleSubscriptions)
}
