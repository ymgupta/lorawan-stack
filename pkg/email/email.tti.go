// Copyright © 2019 The Things Industries B.V.

package email

import (
	"context"
	"net/url"

	"go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/tenant"
)

func (c Config) Apply(ctx context.Context) Config {
	if license.RequireMultiTenancy(ctx) != nil {
		return c
	}
	tenantID := tenant.FromContext(ctx)
	if tenantID.TenantID == "" {
		return c
	}
	deriv := c
	if isURL, err := url.Parse(c.Network.IdentityServerURL); err == nil {
		isURL.Host = tenantID.TenantID + "." + isURL.Host
		deriv.Network.IdentityServerURL = isURL.String()
	}
	if consoleURL, err := url.Parse(c.Network.ConsoleURL); err == nil {
		consoleURL.Host = tenantID.TenantID + "." + consoleURL.Host
		deriv.Network.ConsoleURL = consoleURL.String()
	}
	if tenantFetcher, ok := tenant.FetcherFromContext(ctx); ok {
		if tenant, err := tenantFetcher.FetchTenant(ctx, &tenantID, "name"); err == nil {
			if tenant.Name != "" {
				deriv.Network.Name = tenant.Name
			}
		}
	}
	return deriv
}
