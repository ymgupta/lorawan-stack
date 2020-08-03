// Copyright © 2019 The Things Industries B.V.

package oauth

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/oauth/oidc"
	"go.thethings.network/lorawan-stack/v3/pkg/webui"
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

// GetOIDCConfig returns the OIDC provider configuration.
func (s *server) GetOIDCConfig(ctx context.Context) oidc.Config {
	return s.configProvider(ctx).Apply(ctx).Providers.OIDC
}

// GetTemplateData returns the web template configuration.
func (s *server) GetTemplateData(ctx context.Context) webui.TemplateData {
	return s.configProvider(ctx).Apply(ctx).UI.TemplateData
}
