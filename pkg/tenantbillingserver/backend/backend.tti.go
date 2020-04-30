// Copyright © 2019 The Things Industries B.V.

package backend

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

const (
	// ManagedByTenantAttribute is the backend that manages the tenant.
	ManagedByTenantAttribute = "managed-by"
)

// Interface is an tenant handling backend.
type Interface interface {
	// Report reports the tenant registry totals to the backend. Only the attributes and IDs are retrieved for the tenant.
	Report(ctx context.Context, tenant *ttipb.Tenant, totals *ttipb.TenantRegistryTotals) error
}
