// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/tenant"
)

func (AuthorizationCode) _isMultiTenant() {}

// SetContext needs to be called before creating models.
func (c *AuthorizationCode) SetContext(ctx context.Context) {
	c.TenantID = tenant.FromContext(ctx).TenantID
	c.Model.SetContext(ctx)
}

func (AccessToken) _isMultiTenant() {}

// SetContext needs to be called before creating models.
func (t *AccessToken) SetContext(ctx context.Context) {
	t.TenantID = tenant.FromContext(ctx).TenantID
	t.Model.SetContext(ctx)
}
