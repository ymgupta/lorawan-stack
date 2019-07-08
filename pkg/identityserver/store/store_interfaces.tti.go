// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

// TenantStore interface for storing Tenants.
//
// All functions assume the input and fieldMask to be validated, and assume
// sufficient rights to perform the action.
type TenantStore interface {
	CreateTenant(ctx context.Context, app *ttipb.Tenant) (*ttipb.Tenant, error)
	FindTenants(ctx context.Context, ids []*ttipb.TenantIdentifiers, fieldMask *types.FieldMask) ([]*ttipb.Tenant, error)
	GetTenant(ctx context.Context, id *ttipb.TenantIdentifiers, fieldMask *types.FieldMask) (*ttipb.Tenant, error)
	UpdateTenant(ctx context.Context, app *ttipb.Tenant, fieldMask *types.FieldMask) (*ttipb.Tenant, error)
	DeleteTenant(ctx context.Context, id *ttipb.TenantIdentifiers) error
}
