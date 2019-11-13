// Copyright © 2019 The Things Industries B.V.

package console

import (
	"context"
)

// Apply the context to the config.
func (conf Config) Apply(ctx context.Context) Config {
	deriv := conf
	deriv.OAuth = conf.OAuth.Apply(ctx)
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
	deriv.GS = conf.GS.Apply(ctx)
	deriv.NS = conf.NS.Apply(ctx)
	deriv.AS = conf.AS.Apply(ctx)
	deriv.JS = conf.JS.Apply(ctx)
	deriv.EDTC = conf.EDTC.Apply(ctx)
	deriv.QRG = conf.QRG.Apply(ctx)
	return deriv
}

// Apply the context to the config.
func (conf FrontendConfig) Apply(ctx context.Context) FrontendConfig {
	deriv := conf
	deriv.StackConfig = conf.StackConfig.Apply(ctx)
	return deriv
}
