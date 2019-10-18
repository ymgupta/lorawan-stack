// Copyright © 2019 The Things Industries B.V.

package stripe

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
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

// RegisterRoutes implements web.Registerer.
func (s *Stripe) RegisterRoutes(srv *web.Server) {
	group := srv.Group(ttnpb.HTTPAPIPrefix + "/tbs/stripe")
	group.Any("/subscriptions", s.handleSubscriptions)
}
