// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/tenant"
)

func (User) _isMultiTenant() {}

func (usr *User) SetContext(ctx context.Context) {
	usr.TenantID = tenant.FromContext(ctx).TenantID
	usr.Model.SetContext(ctx)
	usr.Account.SetContext(ctx)
}
