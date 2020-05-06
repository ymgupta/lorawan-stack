// Copyright © 2019 The Things Industries B.V.

package webui

import (
	"context"
	"net/url"

	"go.thethings.network/lorawan-stack/v3/pkg/license"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
)

// Apply the context to the data.
func (t TemplateData) Apply(ctx context.Context) TemplateData {
	if license.RequireMultiTenancy(ctx) != nil {
		return t
	}
	tenantID := tenant.FromContext(ctx)
	if tenantID.TenantID == "" {
		return t
	}
	deriv := t
	if canonical, err := url.Parse(t.CanonicalURL); err == nil {
		canonical.Host = tenantID.TenantID + "." + canonical.Host
		deriv.CanonicalURL = canonical.String()
	}
	if tenantFetcher, ok := tenant.FetcherFromContext(ctx); ok {
		if tenant, err := tenantFetcher.FetchTenant(ctx, &tenantID, "name", "configuration"); err == nil {
			if tenant.Name != "" {
				deriv.SiteName = tenant.Name
			}
			clusterConfiguration := tenant.GetConfiguration().GetDefaultCluster()
			if url := clusterConfiguration.GetUI().GetBrandingBaseURL(); url != "" {
				deriv.BrandingBaseURL = url
			}
		}
	}
	return deriv
}
