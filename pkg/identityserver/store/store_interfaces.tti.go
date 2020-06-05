// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"

	ptypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// TenantStore interface for storing Tenants.
//
// All functions assume the input and fieldMask to be validated, and assume
// sufficient rights to perform the action.
type TenantStore interface {
	CreateTenant(ctx context.Context, tnt *ttipb.Tenant) (*ttipb.Tenant, error)
	FindTenants(ctx context.Context, ids []*ttipb.TenantIdentifiers, fieldMask *ptypes.FieldMask) ([]*ttipb.Tenant, error)
	GetTenant(ctx context.Context, id *ttipb.TenantIdentifiers, fieldMask *ptypes.FieldMask) (*ttipb.Tenant, error)
	UpdateTenant(ctx context.Context, tnt *ttipb.Tenant, fieldMask *ptypes.FieldMask) (*ttipb.Tenant, error)
	DeleteTenant(ctx context.Context, id *ttipb.TenantIdentifiers) error
	GetTenantIDForEndDeviceEUIs(ctx context.Context, joinEUI, devEUI types.EUI64) (*ttipb.TenantIdentifiers, error)
	GetTenantIDForGatewayEUI(ctx context.Context, eui types.EUI64) (*ttipb.TenantIdentifiers, error)
	CountEntities(ctx context.Context, id *ttipb.TenantIdentifiers, entityType string) (uint64, error)
}

// ExternalUserStore interface for storing associations between external users
// and local users.
//
// All functions assume the input and fieldMask to be validated, and assume
// sufficient rights to perform the action.
type ExternalUserStore interface {
	CreateExternalUser(ctx context.Context, eu *ttipb.ExternalUser) (*ttipb.ExternalUser, error)
	GetExternalUserByUserID(ctx context.Context, ids *ttnpb.UserIdentifiers) (*ttipb.ExternalUser, error)
	GetExternalUserByExternalID(ctx context.Context, id string) (*ttipb.ExternalUser, error)
	DeleteExternalUser(ctx context.Context, id string) error
}
