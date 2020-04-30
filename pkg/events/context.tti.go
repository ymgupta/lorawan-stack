// Copyright © 2019 The Things Industries B.V.

package events

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

type tenantIDContextMarshaler struct{}

func (tenantIDContextMarshaler) MarshalContext(ctx context.Context) []byte {
	if id := tenant.FromContext(ctx); !id.IsZero() {
		if b, err := id.Marshal(); err == nil {
			return b
		}
	}
	return nil
}

func (tenantIDContextMarshaler) UnmarshalContext(ctx context.Context, b []byte) (context.Context, error) {
	var id ttipb.TenantIdentifiers
	if err := id.Unmarshal(b); err != nil {
		return nil, err
	}
	return tenant.NewContext(ctx, id), nil
}

func init() {
	RegisterContextMarshaler("tenant-id", tenantIDContextMarshaler{})
}
