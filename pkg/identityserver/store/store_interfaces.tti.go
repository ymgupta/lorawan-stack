// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"

	ptypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// TenantStore interface for storing Tenants.
//
// All functions assume the input and fieldMask to be validated, and assume
// sufficient rights to perform the action.
type TenantStore interface {
	CreateTenant(ctx context.Context, app *ttipb.Tenant) (*ttipb.Tenant, error)
	FindTenants(ctx context.Context, ids []*ttipb.TenantIdentifiers, fieldMask *ptypes.FieldMask) ([]*ttipb.Tenant, error)
	GetTenant(ctx context.Context, id *ttipb.TenantIdentifiers, fieldMask *ptypes.FieldMask) (*ttipb.Tenant, error)
	UpdateTenant(ctx context.Context, app *ttipb.Tenant, fieldMask *ptypes.FieldMask) (*ttipb.Tenant, error)
	DeleteTenant(ctx context.Context, id *ttipb.TenantIdentifiers) error
	GetTenantIDForEndDeviceEUIs(ctx context.Context, joinEUI, devEUI types.EUI64) (*ttipb.TenantIdentifiers, error)
	GetTenantIDForGatewayEUI(ctx context.Context, eui types.EUI64) (*ttipb.TenantIdentifiers, error)
	CountEntities(ctx context.Context, id *ttipb.TenantIdentifiers, entityType string) (uint64, error)
}
