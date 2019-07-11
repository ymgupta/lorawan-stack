// Copyright © 2019 The Things Industries B.V.

package web

import "go.thethings.network/lorawan-stack/pkg/tenant"

// WithTenantConfig adds tenant configuration.
func WithTenantConfig(config tenant.Config) Option {
	return func(o *options) {
		o.tenant = config
	}
}
