// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/tenant"
)

func (APIKey) _isMultiTenant() {}

// SetContext needs to be called before creating models.
func (k *APIKey) SetContext(ctx context.Context) {
	k.TenantID = tenant.FromContext(ctx).TenantID
	k.Model.SetContext(ctx)
}
