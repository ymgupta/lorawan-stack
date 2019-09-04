// Copyright © 2019 The Things Industries B.V.

package identityserver

import (
	"context"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/identityserver/blacklist"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func (is *IdentityServer) createTenant(ctx context.Context, req *ttipb.CreateTenantRequest) (tnt *ttipb.Tenant, err error) {
	if err := license.RequireMultiTenancy(ctx); err != nil {
		return nil, err
	}
	if err = blacklist.Check(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if !tenantRightsFromContext(ctx).admin {
		return nil, errNoTenantRights
	}
	if err := validateContactInfo(req.Tenant.ContactInfo); err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		tnt, err = store.GetTenantStore(db).CreateTenant(ctx, &req.Tenant)
		if err != nil {
			return err
		}
		if len(req.ContactInfo) > 0 {
			cleanContactInfo(req.ContactInfo)
			tnt.ContactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, tnt.TenantIdentifiers, req.ContactInfo)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tnt, nil
}

func (is *IdentityServer) getTenant(ctx context.Context, req *ttipb.GetTenantRequest) (tnt *ttipb.Tenant, err error) {
	if rights := tenantRightsFromContext(ctx); !rights.admin {
		if rights.readCurrent && req.GetTenantID() == tenant.FromContext(ctx).TenantID {
			defer func() { tnt = tnt.PublicSafe() }()
		} else {
			return nil, errNoTenantRights
		}
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttipb.TenantFieldPathsNested, req.FieldMask.Paths, getPaths, nil)
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		tnt, err = store.GetTenantStore(db).GetTenant(ctx, &req.TenantIdentifiers, &req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "contact_info") {
			tnt.ContactInfo, err = store.GetContactInfoStore(db).GetContactInfo(ctx, tnt.TenantIdentifiers)
			if err != nil {
				return err
			}
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return tnt, nil
}

func (is *IdentityServer) listTenants(ctx context.Context, req *ttipb.ListTenantsRequest) (tnts *ttipb.Tenants, err error) {
	if !tenantRightsFromContext(ctx).admin {
		return nil, errNoTenantRights
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttipb.TenantFieldPathsNested, req.FieldMask.Paths, getPaths, nil)
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	tnts = &ttipb.Tenants{}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		tnts.Tenants, err = store.GetTenantStore(db).FindTenants(ctx, nil, &req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tnts, nil
}

func (is *IdentityServer) updateTenant(ctx context.Context, req *ttipb.UpdateTenantRequest) (tnt *ttipb.Tenant, err error) {
	if rights := tenantRightsFromContext(ctx); !rights.admin {
		if rights.readCurrent && req.GetTenantID() == tenant.FromContext(ctx).TenantID {
			if !ttnpb.HasOnlyAllowedFields(req.FieldMask.Paths) {
				return nil, errInsufficientTenantRights
			}
		} else {
			return nil, errNoTenantRights
		}
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttipb.TenantFieldPathsNested, req.FieldMask.Paths, nil, getPaths)
	if len(req.FieldMask.Paths) == 0 {
		req.FieldMask.Paths = updatePaths
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "contact_info") {
		if err := validateContactInfo(req.Tenant.ContactInfo); err != nil {
			return nil, err
		}
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		tnt, err = store.GetTenantStore(db).UpdateTenant(ctx, &req.Tenant, &req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "contact_info") {
			cleanContactInfo(req.ContactInfo)
			tnt.ContactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, tnt.TenantIdentifiers, req.ContactInfo)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tnt, nil
}

func (is *IdentityServer) deleteTenant(ctx context.Context, ids *ttipb.TenantIdentifiers) (*types.Empty, error) {
	if !tenantRightsFromContext(ctx).admin {
		return nil, errNoTenantRights
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		return store.GetTenantStore(db).DeleteTenant(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (is *IdentityServer) getTenantIdentifiersForEndDeviceEUIs(ctx context.Context, req *ttipb.GetTenantIdentifiersForEndDeviceEUIsRequest) (ids *ttipb.TenantIdentifiers, err error) {
	if !tenantRightsFromContext(ctx).readCurrent {
		return nil, errNoTenantRights
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		ids, err = store.GetTenantStore(db).GetTenantIDForEndDeviceEUIs(ctx, req.JoinEUI, req.DevEUI)
		return err
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (is *IdentityServer) getTenantIdentifiersForGatewayEUI(ctx context.Context, req *ttipb.GetTenantIdentifiersForGatewayEUIRequest) (ids *ttipb.TenantIdentifiers, err error) {
	if !tenantRightsFromContext(ctx).readCurrent {
		return nil, errNoTenantRights
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		ids, err = store.GetTenantStore(db).GetTenantIDForGatewayEUI(ctx, req.EUI)
		return err
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (is *IdentityServer) getTenantRegistryTotals(ctx context.Context, req *ttipb.GetTenantRegistryTotalsRequest) (*ttipb.TenantRegistryTotals, error) {
	if !tenantRightsFromContext(ctx).readCurrent {
		return nil, errNoTenantRights
	}
	if len(req.FieldMask.Paths) == 0 {
		req.FieldMask.Paths = ttipb.TenantRegistryTotalsFieldPathsTopLevel
	}
	var totals ttipb.TenantRegistryTotals
	err := is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		store := store.GetTenantStore(db)
		if ttnpb.HasAnyField(req.FieldMask.Paths, "applications") {
			applications, err := store.CountEntities(ctx, req.TenantIdentifiers, "application")
			if err != nil {
				return err
			}
			totals.Applications = applications
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "clients") {
			clients, err := store.CountEntities(ctx, req.TenantIdentifiers, "client")
			if err != nil {
				return err
			}
			totals.Clients = clients
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "end_devices") {
			endDevices, err := store.CountEntities(ctx, req.TenantIdentifiers, "end_device")
			if err != nil {
				return err
			}
			totals.EndDevices = endDevices
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "gateways") {
			gateways, err := store.CountEntities(ctx, req.TenantIdentifiers, "gateway")
			if err != nil {
				return err
			}
			totals.Gateways = gateways
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "organizations") {
			organizations, err := store.CountEntities(ctx, req.TenantIdentifiers, "organization")
			if err != nil {
				return err
			}
			totals.Organizations = organizations
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "users") {
			users, err := store.CountEntities(ctx, req.TenantIdentifiers, "user")
			if err != nil {
				return err
			}
			totals.Users = users
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &totals, nil
}

type tenantRegistry struct {
	*IdentityServer
}

func (tr *tenantRegistry) Create(ctx context.Context, req *ttipb.CreateTenantRequest) (*ttipb.Tenant, error) {
	return tr.createTenant(ctx, req)
}

func (tr *tenantRegistry) Get(ctx context.Context, req *ttipb.GetTenantRequest) (*ttipb.Tenant, error) {
	return tr.getTenant(ctx, req)
}

func (tr *tenantRegistry) List(ctx context.Context, req *ttipb.ListTenantsRequest) (*ttipb.Tenants, error) {
	return tr.listTenants(ctx, req)
}

func (tr *tenantRegistry) Update(ctx context.Context, req *ttipb.UpdateTenantRequest) (*ttipb.Tenant, error) {
	return tr.updateTenant(ctx, req)
}

func (tr *tenantRegistry) Delete(ctx context.Context, req *ttipb.TenantIdentifiers) (*types.Empty, error) {
	return tr.deleteTenant(ctx, req)
}

func (tr *tenantRegistry) GetIdentifiersForEndDeviceEUIs(ctx context.Context, req *ttipb.GetTenantIdentifiersForEndDeviceEUIsRequest) (*ttipb.TenantIdentifiers, error) {
	return tr.getTenantIdentifiersForEndDeviceEUIs(ctx, req)
}

func (tr *tenantRegistry) GetIdentifiersForGatewayEUI(ctx context.Context, req *ttipb.GetTenantIdentifiersForGatewayEUIRequest) (*ttipb.TenantIdentifiers, error) {
	return tr.getTenantIdentifiersForGatewayEUI(ctx, req)
}

func (tr *tenantRegistry) GetRegistryTotals(ctx context.Context, req *ttipb.GetTenantRegistryTotalsRequest) (*ttipb.TenantRegistryTotals, error) {
	return tr.getTenantRegistryTotals(ctx, req)
}
