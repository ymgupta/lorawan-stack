// Copyright © 2019 The Things Industries B.V.

package web

import (
	"context"
	"net/url"

	"go.thethings.network/lorawan-stack/v3/pkg/license"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
)

// Apply the context to the data.
func (c DownlinksConfig) Apply(ctx context.Context) DownlinksConfig {
	if license.RequireMultiTenancy(ctx) != nil {
		return c
	}
	tenantID := tenant.FromContext(ctx).TenantID
	if tenantID == "" {
		return c
	}
	deriv := c
	if c.PublicAddress != "" {
		base, err := url.Parse(c.PublicAddress)
		if err == nil {
			base.Host = tenantID + "." + base.Host
			deriv.PublicAddress = base.String()
		}
	}
	if c.PublicTLSAddress != "" {
		base, err := url.Parse(c.PublicTLSAddress)
		if err == nil {
			base.Host = tenantID + "." + base.Host
			deriv.PublicTLSAddress = base.String()
		}
	}
	return deriv
}
