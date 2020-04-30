// Copyright © 2019 The Things Industries B.V.

package webui

import (
	"context"
	"net/url"

	"go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/tenant"
)

// Apply the context to the config.
func (conf APIConfig) Apply(ctx context.Context) APIConfig {
	if license.RequireMultiTenancy(ctx) != nil {
		return conf
	}
	deriv := conf
	if base, err := url.Parse(conf.BaseURL); err == nil {
		if tenantID := tenant.FromContext(ctx).TenantID; tenantID != "" {
			base.Host = tenantID + "." + base.Host
			deriv.BaseURL = base.String()
		}
	}
	return deriv
}
