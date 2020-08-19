// Copyright © 2020 The Things Industries B.V.

package oidc

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
)

type config struct {
	ID                 string `name:"name" description:"Provider ID"`
	Name               string `name:"name" description:"Public provider name"`
	AllowRegistrations bool   `name:"allow-registrations" description:"Allow clients to be registered automatically on login"`
	ClientID           string `name:"client-id" description:"Client ID of the OIDC client"`
	ClientSecret       string `name:"client-secret" description:"Client secret of the OIDC client"`
	ProviderURL        string `name:"provider-url" description:"Path of the OIDC server"`
}

var errWrongProvider = errors.DefineInvalidArgument("wrong_provider", "wrong provider")

type providerConfigKeyType struct{}

var providerConfigKey providerConfigKeyType

// WithProvider attaches the authentication provider configuration to the context.
func WithProvider(ctx context.Context, pb *ttipb.AuthenticationProvider) (context.Context, error) {
	oidc := pb.GetConfiguration().GetOIDC()
	if oidc == nil {
		return nil, errWrongProvider.New()
	}
	return context.WithValue(ctx, providerConfigKey, &config{
		ID:                 pb.ProviderID,
		Name:               pb.Name,
		AllowRegistrations: pb.AllowRegistrations,
		ClientID:           oidc.ClientID,
		ClientSecret:       oidc.ClientSecret,
		ProviderURL:        oidc.ProviderURL,
	}), nil
}

func configFromContext(ctx context.Context) *config {
	if c, ok := ctx.Value(providerConfigKey).(*config); ok {
		return c
	}
	return &config{}
}
