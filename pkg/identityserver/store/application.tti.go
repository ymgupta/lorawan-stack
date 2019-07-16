// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/tenant"
)

func (Application) _isMultiTenant() {}

// SetContext needs to be called before creating models.
func (app *Application) SetContext(ctx context.Context) {
	app.TenantID = tenant.FromContext(ctx).TenantID
	app.Model.SetContext(ctx)
}
