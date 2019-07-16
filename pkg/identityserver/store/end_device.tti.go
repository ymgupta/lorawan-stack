// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/tenant"
)

func (EndDevice) _isMultiTenant() {}

// SetContext needs to be called before creating models.
func (dev *EndDevice) SetContext(ctx context.Context) {
	dev.TenantID = tenant.FromContext(ctx).TenantID
	dev.Model.SetContext(ctx)
}
