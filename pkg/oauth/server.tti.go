// Copyright © 2019 The Things Industries B.V.

package oauth

import (
	"context"
	"net/url"

	"go.thethings.network/lorawan-stack/v3/pkg/license"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
)

// ConfigurationPatcher is a configuration patcher for the OAuth configuration.
type ConfigurationPatcher interface {
	Apply(context.Context, Config) Config
}

// ConfigurationPatcherFunc is a ConfigurationPatcher in functional form.
type ConfigurationPatcherFunc func(context.Context, Config) Config

// Apply patches the configuration using the function.
func (f ConfigurationPatcherFunc) Apply(ctx context.Context, conf Config) Config {
	return f(ctx, conf)
}

type configPatchKeyType struct{}

var configPatchKey configPatchKeyType

// WithConfigurationPatcher adds the configuration patcher to the context.
func WithConfigurationPatcher(ctx context.Context, patcher ConfigurationPatcher) context.Context {
	return context.WithValue(ctx, configPatchKey, patcher)
}

func configurationPatcherFromContext(ctx context.Context) ConfigurationPatcher {
	if patcher, ok := ctx.Value(configPatchKey).(ConfigurationPatcher); ok {
		return patcher
	}
	return nil
}

// Apply the context to the config.
func (conf Config) Apply(ctx context.Context) Config {
	deriv := conf
	if patcher := configurationPatcherFromContext(ctx); patcher != nil {
		deriv = patcher.Apply(ctx, deriv)
	}
	deriv.UI = conf.UI.Apply(ctx)
	deriv.Providers = conf.Providers.Apply(ctx)
	return deriv
}

// Apply the context to the config.
func (conf ProvidersConfig) Apply(ctx context.Context) ProvidersConfig {
	deriv := conf
	deriv.OIDC = conf.OIDC.Apply(ctx)
	return deriv
}

// Apply the context to the config.
func (conf OIDCConfig) Apply(ctx context.Context) OIDCConfig {
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

// Apply the context to the config.
func (conf UIConfig) Apply(ctx context.Context) UIConfig {
	deriv := conf
	deriv.TemplateData = conf.TemplateData.Apply(ctx)
	deriv.FrontendConfig = conf.FrontendConfig.Apply(ctx)
	return deriv
}

// Apply the context to the config.
func (conf StackConfig) Apply(ctx context.Context) StackConfig {
	deriv := conf
	deriv.IS = conf.IS.Apply(ctx)
	return deriv
}

// Apply the context to the config.
func (conf FrontendConfig) Apply(ctx context.Context) FrontendConfig {
	deriv := conf
	deriv.StackConfig = conf.StackConfig.Apply(ctx)
	return deriv
}
