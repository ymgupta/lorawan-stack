// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/tenant"
)

func (Gateway) _isMultiTenant() {}

// SetContext needs to be called before creating models.
func (gtw *Gateway) SetContext(ctx context.Context) {
	gtw.TenantID = tenant.FromContext(ctx).TenantID
	gtw.Model.SetContext(ctx)
}
