// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/tenant"
)

func (Client) _isMultiTenant() {}

// SetContext needs to be called before creating models.
func (cli *Client) SetContext(ctx context.Context) {
	cli.TenantID = tenant.FromContext(ctx).TenantID
	cli.Model.SetContext(ctx)
}
