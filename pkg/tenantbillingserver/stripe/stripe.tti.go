// Copyright © 2019 The Things Industries B.V.

package stripe

import (
	"context"
	"time"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/client"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/tenantbillingserver/backend"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/web"
	"google.golang.org/grpc"
)

// Config is the configuration for the Stripe payment backend.
type Config struct {
	Enable                  bool     `name:"enable" description:"Enable the Stripe backend"`
	APIKey                  string   `name:"api-key" description:"API key used to connect to Stripe"`
	EndpointSecretKey       string   `name:"endpoint-secret-key" description:"Endpoint secret key used to verify webhook signatures"`
	SkipSignatureValidation bool     `name:"skip-signature-validation" description:"Skip the webhook signature validation"`
	RecurringPlanIDs        []string `name:"recurring-plan-ids" description:"Recurring pricing plan IDs to be handled"`
	MeteredPlanIDs          []string `name:"metered-plan-ids" description:"Metered pricing plan IDs to be handled"`
}

var (
	errNoAPIKey            = errors.DefineInvalidArgument("no_api_key", "no API key provided")
	errNoEndpointSecretKey = errors.DefineInvalidArgument("no_endpoint_secret_key", "no endpoint secret key provided")
	errNoPlanIDs           = errors.DefineInvalidArgument("no_plan_ids", "no plan IDs provided")
	errNoTenantAttribute   = errors.DefineInvalidArgument("no_tenant_attribute", "no tenant attribute `{attribute}` available")
	errUnknownTenantState  = errors.DefineInternal("unknown_tenant_state", "tenant state `{state}` is unknown")
)

// New returns a new Stripe backend using the config.
func (c Config) New(ctx context.Context, component *component.Component, opts ...Option) (*Stripe, error) {
	if c.APIKey == "" {
		return nil, errNoAPIKey.New()
	}
	if c.EndpointSecretKey == "" && !c.SkipSignatureValidation {
		return nil, errNoEndpointSecretKey.New()
	}
	if len(c.RecurringPlanIDs)+len(c.MeteredPlanIDs) == 0 {
		return nil, errNoPlanIDs.New()
	}
	return New(ctx, component, c, opts...)
}

// Stripe is the payment and tenant configuration backend.
type Stripe struct {
	ctx       context.Context
	component *component.Component
	config    Config

	recurringPlanIDs map[string]struct{}
	meteredPlanIDs   map[string]struct{}

	tenantsClient ttipb.TenantRegistryClient
	tenantAuth    grpc.CallOption

	apiClient *client.API
}

// New returns a new Stripe backend.
func New(ctx context.Context, component *component.Component, config Config, opts ...Option) (*Stripe, error) {
	s := &Stripe{
		ctx:              log.NewContextWithField(ctx, "namespace", "tenantbillingserver/stripe"),
		component:        component,
		config:           config,
		recurringPlanIDs: make(map[string]struct{}),
		meteredPlanIDs:   make(map[string]struct{}),
		tenantAuth:       grpc.EmptyCallOption{},
	}
	for _, opt := range opts {
		opt(s)
	}
	for _, v := range config.RecurringPlanIDs {
		s.recurringPlanIDs[v] = struct{}{}
	}
	for _, v := range config.MeteredPlanIDs {
		s.meteredPlanIDs[v] = struct{}{}
	}
	return s, nil
}

// Option is an option for the Stripe backend.
type Option func(*Stripe)

// WithTenantRegistryClient sets the backend to use the given tenant registry client.
func WithTenantRegistryClient(client ttipb.TenantRegistryClient) Option {
	return Option(func(s *Stripe) {
		s.tenantsClient = client
	})
}

// WithTenantRegistryAuth sets the backend to use the given tenant registry authentication.
func WithTenantRegistryAuth(auth grpc.CallOption) Option {
	return Option(func(s *Stripe) {
		s.tenantAuth = auth
	})
}

// WithStripeAPIClient sets the backend to use the given Stripe API client. Generally used for testing.
func WithStripeAPIClient(c *client.API) Option {
	return Option(func(s *Stripe) {
		s.apiClient = c
	})
}

// Report updates the Stripe subscription of the tenant if the tenant is managed by Stripe.
func (s *Stripe) Report(ctx context.Context, tenant *ttipb.Tenant, totals *ttipb.TenantRegistryTotals) error {
	manager, ok := tenant.Attributes[backend.ManagedByTenantAttribute]
	if !ok || manager != managerAttributeValue {
		return nil
	}

	switch tenant.State {
	case ttnpb.STATE_REQUESTED, ttnpb.STATE_REJECTED, ttnpb.STATE_SUSPENDED:
		// Do not report metrics of inactive tenants.
		return nil
	case ttnpb.STATE_APPROVED, ttnpb.STATE_FLAGGED:
		break
	default:
		return errUnknownTenantState.WithAttributes("state", tenant.State)
	}

	planID := tenant.Attributes[planIDAttribute]
	if planID == "" {
		return errNoTenantAttribute.WithAttributes("attribute", planIDAttribute)
	}
	customerID := tenant.Attributes[customerIDAttribute]
	if customerID == "" {
		return errNoTenantAttribute.WithAttributes("attribute", customerIDAttribute)
	}
	subscriptionID := tenant.Attributes[subscriptionIDAttribute]
	if subscriptionID == "" {
		return errNoTenantAttribute.WithAttributes("attribute", subscriptionIDAttribute)
	}
	subscriptionItemID := tenant.Attributes[subscriptionItemIDAttribute]
	if subscriptionItemID == "" {
		return errNoTenantAttribute.WithAttributes("attribute", subscriptionItemIDAttribute)
	}

	quantity := int64(totals.GetEndDevices())
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"tenant_id", tenant.TenantID, // The context does not contain a tenant.
		"plan_id", planID,
		"customer_id", customerID,
		"subscription_id", subscriptionID,
		"subscription_item_id", subscriptionItemID,
		"quantity", quantity,
	))

	if _, ok := s.recurringPlanIDs[planID]; ok {
		// Recurring plans do not need usage records.
	} else if _, ok := s.meteredPlanIDs[planID]; ok {
		params := &stripe.UsageRecordParams{
			Quantity:         stripe.Int64(quantity),
			Timestamp:        stripe.Int64(time.Now().UTC().Unix()),
			SubscriptionItem: stripe.String(subscriptionItemID),
			Action:           stripe.String(stripe.UsageRecordActionSet),
		}
		if _, err := s.newUsageRecord(params); err != nil {
			return err
		}
		logger.Debug("Usage recorded")
	} else {
		logger.Error("Unrecognized plan ID")
	}
	return nil
}

// RegisterRoutes implements web.Registerer.
func (s *Stripe) RegisterRoutes(srv *web.Server) {
	srv.POST(ttnpb.HTTPAPIPrefix+"/tbs/stripe", s.handleWebhook)
}

func (s *Stripe) getAPIClient() (*client.API, error) {
	if s.apiClient != nil {
		return s.apiClient, nil
	}
	backends := stripe.NewBackends(nil)
	backends.API = stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
		LeveledLogger: log.FromContext(s.ctx),
		LogLevel:      convertLogLevel(s.ctx, s.component),
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

func (s *Stripe) newUsageRecord(params *stripe.UsageRecordParams) (*stripe.UsageRecord, error) {
	client, err := s.getAPIClient()
	if err != nil {
		return nil, err
	}
	return client.UsageRecords.New(params)
}

func convertLogLevel(ctx context.Context, c *component.Component) int {
	level := c.GetBaseConfig(ctx).Log.Level
	switch level {
	case log.DebugLevel:
		return 3
	case log.InfoLevel:
		return 2
	case log.WarnLevel, log.ErrorLevel, log.FatalLevel:
		return 1
	default:
		log.FromContext(ctx).WithField("level", level).Error("Unknown log level provided. Default to 0")
		return 0
	}
}
