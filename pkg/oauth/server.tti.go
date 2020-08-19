// Copyright © 2019 The Things Industries B.V.

package oauth

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/webui"
)

// Apply the context to the config.
func (conf Config) Apply(ctx context.Context) Config {
	deriv := conf
	deriv.UI = conf.UI.Apply(ctx)
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

// GetTemplateData returns the web template configuration.
func (s *server) GetTemplateData(ctx context.Context) webui.TemplateData {
	return s.configFromContext(ctx).UI.TemplateData
}
