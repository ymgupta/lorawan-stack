// Copyright © 2019 The Things Industries B.V.

package oauthclient

import (
	"context"
	"net/url"

	"go.thethings.network/lorawan-stack/v3/pkg/license"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
)

// Apply the context to the config.
func (conf Config) Apply(ctx context.Context) Config {
	if license.RequireMultiTenancy(ctx) != nil {
		return conf
	}
	deriv := conf
	if auth, err := url.Parse(conf.AuthorizeURL); err == nil {
		if tenantID := tenant.FromContext(ctx).TenantID; tenantID != "" {
			auth.Host = tenantID + "." + auth.Host
			deriv.AuthorizeURL = auth.String()
		}
	}
	if token, err := url.Parse(conf.TokenURL); err == nil {
		if tenantID := tenant.FromContext(ctx).TenantID; tenantID != "" {
			token.Host = tenantID + "." + token.Host
			deriv.TokenURL = token.String()
		}
	}
	if root, err := url.Parse(conf.RootURL); err == nil {
		if tenantID := tenant.FromContext(ctx).TenantID; tenantID != "" {
			root.Host = tenantID + "." + root.Host
			deriv.RootURL = root.String()
		}
	}
	return deriv
}
