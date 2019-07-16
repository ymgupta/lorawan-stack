// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/tenant"
)

func (Invitation) _isMultiTenant() {}

// SetContext needs to be called before creating models.
func (i *Invitation) SetContext(ctx context.Context) {
	i.TenantID = tenant.FromContext(ctx).TenantID
	i.Model.SetContext(ctx)
}
