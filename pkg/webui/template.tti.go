// Copyright © 2019 The Things Industries B.V.

package webui

import (
	"context"
	"net/url"

	"go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/tenant"
)

// Apply the context to the data.
func (t TemplateData) Apply(ctx context.Context) TemplateData {
	if license.RequireMultiTenancy(ctx) != nil {
		return t
	}
	deriv := t
	if canonical, err := url.Parse(t.CanonicalURL); err == nil {
		if tenantID := tenant.FromContext(ctx).TenantID; tenantID != "" {
			canonical.Host = tenantID + "." + canonical.Host
			deriv.CanonicalURL = canonical.String()
		}
	}
	return deriv
}
