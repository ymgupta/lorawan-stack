// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/tenant"
)

func (Account) _isMultiTenant() {}

// SetContext needs to be called before creating models.
func (a *Account) SetContext(ctx context.Context) {
	a.TenantID = tenant.FromContext(ctx).TenantID
	a.Model.SetContext(ctx)
}
