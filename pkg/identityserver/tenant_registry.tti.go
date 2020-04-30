// Copyright © 2019 The Things Industries B.V.

package identityserver

import (
	"context"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/email"
	"go.thethings.network/lorawan-stack/pkg/identityserver/blacklist"
	"go.thethings.network/lorawan-stack/pkg/identityserver/emails"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/validate"
)

func (is *IdentityServer) createTenant(ctx context.Context, req *ttipb.CreateTenantRequest) (tnt *ttipb.Tenant, err error) {
	if err := license.RequireMultiTenancy(ctx); err != nil {
		return nil, err
	}
	if err = blacklist.Check(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if !tenantRightsFromContext(ctx).admin {
		return nil, errNoTenantRights.New()
	}
	if err := validateContactInfo(req.Tenant.ContactInfo); err != nil {
		return nil, err
	}

	var initialUserPassword string
	if req.InitialUser != nil {
		if err = blacklist.Check(ctx, req.InitialUser.UserID); err != nil {
			return nil, err
		}
		if err := validate.Email(req.InitialUser.PrimaryEmailAddress); err != nil {
			return nil, err
		}
		if err := validateContactInfo(req.InitialUser.ContactInfo); err != nil {
			return nil, err
		}

		var primaryEmailAddressFound bool
		for _, contactInfo := range req.InitialUser.ContactInfo {
			if contactInfo.ContactMethod == ttnpb.CONTACT_METHOD_EMAIL && contactInfo.Value == req.InitialUser.PrimaryEmailAddress {
				primaryEmailAddressFound = true
				contactInfo.ValidatedAt = req.InitialUser.PrimaryEmailAddressValidatedAt
			}
		}
		if !primaryEmailAddressFound {
			req.InitialUser.ContactInfo = append(req.InitialUser.ContactInfo, &ttnpb.ContactInfo{
				ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
				Value:         req.InitialUser.PrimaryEmailAddress,
				ValidatedAt:   req.InitialUser.PrimaryEmailAddressValidatedAt,
			})
		}
		for _, contactInfo := range req.ContactInfo {
			if contactInfo.ContactMethod == ttnpb.CONTACT_METHOD_EMAIL && contactInfo.Value == req.InitialUser.PrimaryEmailAddress {
				contactInfo.ValidatedAt = req.InitialUser.PrimaryEmailAddressValidatedAt
			}
		}

		if req.InitialUser.Password == "" {
			initialUserPassword, err = auth.GenerateKey(ctx)
			if err != nil {
				return nil, err
			}
			req.InitialUser.Password = initialUserPassword
		}
		hashedPassword, err := auth.Hash(ctx, req.InitialUser.Password)
		if err != nil {
			return nil, err
		}
		req.InitialUser.Password = hashedPassword
		now := time.Now()
		req.InitialUser.PasswordUpdatedAt = &now
	}

	var usr *ttnpb.User

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		tnt, err = store.GetTenantStore(db).CreateTenant(ctx, &req.Tenant)
		if err != nil {
			return err
		}
		if len(req.ContactInfo) > 0 {
			tnt.ContactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, tnt.TenantIdentifiers, req.ContactInfo)
			if err != nil {
				return err
			}
		}
		if req.InitialUser != nil {
			ctx := tenant.NewContext(ctx, tnt.TenantIdentifiers)
			ctx = store.WithoutTenantFetcher(ctx)
			usr, err = store.GetUserStore(db).CreateUser(ctx, req.InitialUser)
			if err != nil {
				return err
			}

			if len(req.InitialUser.ContactInfo) > 0 {
				usr.ContactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, usr.UserIdentifiers, req.InitialUser.ContactInfo)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if usr != nil {
		ctx := tenant.NewContext(ctx, tnt.TenantIdentifiers)
		err = is.SendEmail(ctx, func(data emails.Data) email.MessageData {
			data.SetUser(usr)
			data.Entity.Type, data.Entity.ID = "tenant", tnt.TenantID
			return &emails.TenantCreated{
				Data:              data,
				GlobalNetworkName: is.config.Email.Network.Name,
				InitialPassword:   initialUserPassword,
			}
		})
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Could not send tenant created email")
		}
		_, err := is.requestContactInfoValidation(ctx, usr.EntityIdentifiers())
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Could not send contact info validations")
		}
	}

	return tnt, nil
}

func (is *IdentityServer) getTenant(ctx context.Context, req *ttipb.GetTenantRequest) (tnt *ttipb.Tenant, err error) {
	if !tenantRightsFromContext(ctx).canRead(ctx, &req.TenantIdentifiers) {
		if ttnpb.HasOnlyAllowedFields(req.FieldMask.Paths, ttipb.PublicTenantFields...) {
			defer func() { tnt = tnt.PublicSafe() }()
		} else {
			return nil, errNoTenantRights.New()
		}
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttipb.TenantFieldPathsNested, req.FieldMask.Paths, getPaths, nil)
	err = is.withReadDatabase(ctx, func(db *gorm.DB) (err error) {
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

func (is *IdentityServer) getTenantForFetcher(ctx context.Context, ids *ttipb.TenantIdentifiers, fieldPaths ...string) (tnt *ttipb.Tenant, err error) {
	fieldPaths = cleanFieldMaskPaths(ttipb.TenantFieldPathsNested, fieldPaths, getPaths, nil)
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		tnt, err = store.GetTenantStore(db).GetTenant(ctx, ids, &types.FieldMask{Paths: fieldPaths})
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(fieldPaths, "contact_info") {
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
	if !tenantRightsFromContext(ctx).read {
		return nil, errNoTenantRights.New()
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttipb.TenantFieldPathsNested, req.FieldMask.Paths, getPaths, nil)
	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	tnts = &ttipb.Tenants{}
	err = is.withReadDatabase(ctx, func(db *gorm.DB) (err error) {
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
		if rights.canWrite(ctx, &req.TenantIdentifiers) {
			if !ttnpb.HasOnlyAllowedFields(req.FieldMask.Paths, "configuration") {
				return nil, errInsufficientTenantRights.New()
			}
		} else {
			return nil, errNoTenantRights.New()
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
		return nil, errNoTenantRights.New()
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
	if rights := tenantRightsFromContext(ctx); !rights.read {
		return nil, errNoTenantRights.New()
	}
	err = is.withReadDatabase(ctx, func(db *gorm.DB) (err error) {
		ids, err = store.GetTenantStore(db).GetTenantIDForEndDeviceEUIs(ctx, req.JoinEUI, req.DevEUI)
		return err
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (is *IdentityServer) getTenantIdentifiersForGatewayEUI(ctx context.Context, req *ttipb.GetTenantIdentifiersForGatewayEUIRequest) (ids *ttipb.TenantIdentifiers, err error) {
	if rights := tenantRightsFromContext(ctx); !rights.read {
		return nil, errNoTenantRights.New()
	}
	err = is.withReadDatabase(ctx, func(db *gorm.DB) (err error) {
		ids, err = store.GetTenantStore(db).GetTenantIDForGatewayEUI(ctx, req.EUI)
		return err
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (is *IdentityServer) getTenantRegistryTotals(ctx context.Context, req *ttipb.GetTenantRegistryTotalsRequest) (*ttipb.TenantRegistryTotals, error) {
	if rights := tenantRightsFromContext(ctx); !rights.canRead(ctx, req.TenantIdentifiers) {
		return nil, errNoTenantRights.New()
	}
	if len(req.FieldMask.Paths) == 0 {
		req.FieldMask.Paths = ttipb.TenantRegistryTotalsFieldPathsTopLevel
	}
	var totals ttipb.TenantRegistryTotals
	err := is.withReadDatabase(ctx, func(db *gorm.DB) (err error) {
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
