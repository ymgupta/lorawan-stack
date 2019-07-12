// Copyright © 2019 The Things Industries B.V.

package console

import (
	"context"
	"net/url"

	"go.thethings.network/lorawan-stack/pkg/tenant"
)

// Apply the context to the config.
func (conf Config) Apply(ctx context.Context) Config {
	deriv := conf
	deriv.OAuth = conf.OAuth.Apply(ctx)
	deriv.UI = conf.UI.Apply(ctx)
	return deriv
}

// Apply the context to the config.
func (conf OAuth) Apply(ctx context.Context) OAuth {
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
	return deriv
}

// Apply the context to the config.
func (conf UIConfig) Apply(ctx context.Context) UIConfig {
	deriv := conf
	deriv.TemplateData = conf.TemplateData.Apply(ctx)
	deriv.FrontendConfig = conf.FrontendConfig.Apply(ctx)
	return deriv
}

// Apply the context to the config.
func (conf FrontendConfig) Apply(ctx context.Context) FrontendConfig {
	deriv := conf
	deriv.IS = conf.IS.Apply(ctx)
	deriv.GS = conf.GS.Apply(ctx)
	deriv.NS = conf.NS.Apply(ctx)
	deriv.AS = conf.AS.Apply(ctx)
	deriv.JS = conf.JS.Apply(ctx)
	return deriv
}

// Apply the context to the config.
func (conf APIConfig) Apply(ctx context.Context) APIConfig {
	deriv := conf
	if base, err := url.Parse(conf.BaseURL); err == nil {
		if tenantID := tenant.FromContext(ctx).TenantID; tenantID != "" {
			base.Host = tenantID + "." + base.Host
			deriv.BaseURL = base.String()
		}
	}
	return deriv
}
