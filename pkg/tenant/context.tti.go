// Copyright © 2019 The Things Industries B.V.

package tenant

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

type tenantIDKeyType struct{}

var tenantIDKey = tenantIDKeyType{}

// FromContext returns the current TenantIdentifier based on the given context.
// Returns empty identifier if not found.
func FromContext(ctx context.Context) ttipb.TenantIdentifiers {
	if id, ok := ctx.Value(tenantIDKey).(ttipb.TenantIdentifiers); ok { // set by NewContext
		return id
	}
	return ttipb.TenantIdentifiers{}
}

// NewContext returns a context containing the tenant identifier.
func NewContext(parent context.Context, id ttipb.TenantIdentifiers) context.Context {
	return context.WithValue(parent, tenantIDKey, id)
}
