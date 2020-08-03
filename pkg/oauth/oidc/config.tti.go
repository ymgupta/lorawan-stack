// Copyright © 2020 The Things Industries B.V.

package oidc

import (
	"context"
	"net/url"

	"go.thethings.network/lorawan-stack/v3/pkg/license"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
)

// Config is the configuration of the OpenID Connect authentication provider.
type Config struct {
	Name               string `name:"name" description:"Public provider name"`
	AllowRegistrations bool   `name:"allow-registrations" description:"Allow clients to be registered automatically on login"`
	ClientID           string `name:"client-id" description:"Client ID of the OIDC client"`
	ClientSecret       string `name:"client-secret" description:"Client secret of the OIDC client"`
	RedirectURL        string `name:"redirect-url" description:"Path on the server where the OIDC client will be redirected"`
	ProviderURL        string `name:"provider-url" description:"Path of the OIDC server"`
}

// IsZero checks if the config is empty.
func (conf Config) IsZero() bool {
	return conf.Name == "" &&
		!conf.AllowRegistrations &&
		conf.ClientID == "" &&
		conf.ClientSecret == "" &&
		conf.RedirectURL == "" &&
		conf.ProviderURL == ""
}

// Apply the context to the config.
func (conf Config) Apply(ctx context.Context) Config {
	if license.RequireMultiTenancy(ctx) != nil {
		return conf
	}
	tenantID := tenant.FromContext(ctx)
	if tenantID.TenantID == "" {
		return conf
	}
	deriv := conf
	if redirect, err := url.Parse(conf.RedirectURL); err == nil {
		redirect.Host = tenantID.TenantID + "." + redirect.Host
		deriv.RedirectURL = redirect.String()
	}
	return deriv
}
