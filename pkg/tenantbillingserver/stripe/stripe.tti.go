// Copyright © 2019 The Things Industries B.V.

package stripe

import (
	"context"

	"github.com/stripe/stripe-go/client"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/web"
	"google.golang.org/grpc"
)

// Config is the configuration for the Stripe payment backend.
type Config struct {
	Enabled           bool     `name:"enabled" description:"Enable the backend"`
	APIKey            string   `name:"api-key" description:"The Stripe API key used to report the metrics"`
	EndpointSecretKey string   `name:"endpoint-secret-key" description:"The Stripe endopoint secret used to verify webhook signatures"`
	PlanIDs           []string `name:"plan-ids" description:"The IDs of the subscription plans to be handled"`
	LogLevel          int      `name:"log-level" description:"Log level for the Stripe API client"`
}

var (
	errNoAPIKey  = errors.DefineInvalidArgument("no_api_key", "no API key provided")
	errNoPlanIDs = errors.DefineInvalidArgument("no_plan_ids", "no plan ids provided")
)

// New returns a new Stripe backend using the config.
func (c Config) New(ctx context.Context, component *component.Component, opts ...Option) (*Stripe, error) {
	if !c.Enabled {
		return nil, nil
	}
	if len(c.APIKey) == 0 {
		return nil, errNoAPIKey
	}
	if len(c.PlanIDs) == 0 {
		return nil, errNoPlanIDs
	}
	return New(ctx, component, &c, opts...)
}

// Stripe is the payment and tenant configuration backend.
type Stripe struct {
	ctx       context.Context
	component *component.Component
	config    *Config

	tenantsClient ttipb.TenantRegistryClient
	tenantAuth    grpc.CallOption

	apiClient *client.API
}

// New returns a new Stripe backend.
func New(ctx context.Context, component *component.Component, config *Config, opts ...Option) (*Stripe, error) {
	s := &Stripe{
		ctx:        log.NewContextWithField(ctx, "namespace", "tenantbillingserver/stripe"),
		component:  component,
		config:     config,
		tenantAuth: grpc.EmptyCallOption{},
	}
	for _, opt := range opts {
		opt(s)
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
	return nil
}

// RegisterRoutes implements web.Registerer.
func (s *Stripe) RegisterRoutes(srv *web.Server) {
	srv.POST(ttnpb.HTTPAPIPrefix+"/tbs/stripe", s.handleWebhook)
}
