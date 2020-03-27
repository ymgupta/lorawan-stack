// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

type skipTenantFetcherKeyType struct{}

var skipTenantFetcherKey skipTenantFetcherKeyType

// WithoutTenantFetcher informs the store to query the database directly
// instead of retrieving the tenants using the contextual tenant fetcher.
func WithoutTenantFetcher(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipTenantFetcherKey, true)
}

func skipTenantFetcher(ctx context.Context) bool {
	if skip, ok := ctx.Value(skipTenantFetcherKey).(bool); ok {
		return skip
	}
	return false
}

func checkLicense(model *Model) (*ttipb.License, error) {
	l := license.FromContext(model.ctx)
	if err := license.CheckValidity(&l); err != nil {
		return nil, err
	}
	if err := license.CheckLimitedFunctionality(&l); err != nil {
		return nil, err
	}
	return &l, nil
}

func retrieveTenant(db *gorm.DB, model *Model) (*ttipb.Tenant, error) {
	ctx, tenantID := model.ctx, tenant.FromContext(model.ctx)
	tenantFetcher, ok := tenant.FetcherFromContext(ctx)
	if !ok || skipTenantFetcher(ctx) {
		return GetTenantStore(db).GetTenant(ctx, &tenantID, &types.FieldMask{Paths: entityQuotasFields})
	}
	return tenantFetcher.FetchTenant(ctx, &tenantID, entityQuotasFields...)
}

func countInTenant(ctx context.Context, db *gorm.DB, entityType string) (uint64, error) {
	var count uint64
	query := (&store{db}).query(ctx, nil, withTenantID(tenant.FromContext(ctx).TenantID))
	switch entityType {
	case "organization", "user":
		query = query.Model(Account{}).Where(Account{AccountType: entityType})
	default:
		query = query.Model(modelForEntityType(entityType))
	}
	err := query.Count(&count).Error
	return count, err
}

var errEntityQuotaReached = errors.DefineResourceExhausted("entity_quota", "quota for {entity_type} entities reached")

// BeforeCreate checks if the create is allowed by the license.
func (app *Application) BeforeCreate(db *gorm.DB) error {
	license, err := checkLicense(&app.Model)
	if err != nil {
		return err
	}
	var maxApplications *types.UInt64Value
	if license.MaxApplications != nil {
		maxApplications = license.MaxApplications
	} else if license.MultiTenancy {
		tenant, err := retrieveTenant(db, &app.Model)
		if err != nil {
			return err
		}
		if tenant.MaxApplications != nil {
			maxApplications = tenant.MaxApplications
		}
	}
	if maxApplications != nil {
		if maxApplications.Value == 0 {
			return errEntityQuotaReached.WithAttributes("entity_type", "application")
		}
		count, err := countInTenant(app.Model.ctx, db, "application")
		if err != nil {
			return err
		}
		if count >= maxApplications.Value {
			return errEntityQuotaReached.WithAttributes("entity_type", "application")
		}
	}
	return nil
}

// BeforeCreate checks if the create is allowed by the license.
func (cli *Client) BeforeCreate(db *gorm.DB) error {
	license, err := checkLicense(&cli.Model)
	if err != nil {
		return err
	}
	var maxClients *types.UInt64Value
	if license.MaxClients != nil {
		maxClients = license.MaxClients
	} else if license.MultiTenancy && cli.TenantID != nil {
		tenant, err := retrieveTenant(db, &cli.Model)
		if err != nil {
			return err
		}
		if tenant.MaxClients != nil {
			maxClients = tenant.MaxClients
		}
	}
	if maxClients != nil {
		if maxClients.Value == 0 {
			return errEntityQuotaReached.WithAttributes("entity_type", "client")
		}
		count, err := countInTenant(cli.Model.ctx, db, "client")
		if err != nil {
			return err
		}
		if count >= maxClients.Value {
			return errEntityQuotaReached.WithAttributes("entity_type", "client")
		}
	}
	return nil
}

// BeforeCreate checks if the create is allowed by the license.
func (dev *EndDevice) BeforeCreate(db *gorm.DB) error {
	license, err := checkLicense(&dev.Model)
	if err != nil {
		return err
	}
	var maxEndDevices *types.UInt64Value
	if license.MaxEndDevices != nil {
		maxEndDevices = license.MaxEndDevices
	} else if license.MultiTenancy {
		tenant, err := retrieveTenant(db, &dev.Model)
		if err != nil {
			return err
		}
		if tenant.MaxEndDevices != nil {
			maxEndDevices = tenant.MaxEndDevices
		}
	}
	if maxEndDevices != nil {
		if maxEndDevices.Value == 0 {
			return errEntityQuotaReached.WithAttributes("entity_type", "end device")
		}
		count, err := countInTenant(dev.Model.ctx, db, "end_device")
		if err != nil {
			return err
		}
		if count >= maxEndDevices.Value {
			return errEntityQuotaReached.WithAttributes("entity_type", "end device")
		}
	}
	return nil
}

// BeforeCreate checks if the create is allowed by the license.
func (gtw *Gateway) BeforeCreate(db *gorm.DB) error {
	license, err := checkLicense(&gtw.Model)
	if err != nil {
		return err
	}
	var maxGateways *types.UInt64Value
	if license.MaxGateways != nil {
		maxGateways = license.MaxGateways
	} else if license.MultiTenancy {
		tenant, err := retrieveTenant(db, &gtw.Model)
		if err != nil {
			return err
		}
		if tenant.MaxGateways != nil {
			maxGateways = tenant.MaxGateways
		}
	}
	if maxGateways != nil {
		if maxGateways.Value == 0 {
			return errEntityQuotaReached.WithAttributes("entity_type", "gateway")
		}
		count, err := countInTenant(gtw.Model.ctx, db, "gateway")
		if err != nil {
			return err
		}
		if count >= maxGateways.Value {
			return errEntityQuotaReached.WithAttributes("entity_type", "gateway")
		}
	}
	return nil
}

// BeforeCreate checks if the create is allowed by the license.
func (org *Organization) BeforeCreate(db *gorm.DB) error {
	license, err := checkLicense(&org.Model)
	if err != nil {
		return err
	}
	var maxOrganizations *types.UInt64Value
	if license.MaxOrganizations != nil {
		maxOrganizations = license.MaxOrganizations
	} else if license.MultiTenancy {
		tenant, err := retrieveTenant(db, &org.Model)
		if err != nil {
			return err
		}
		if tenant.MaxOrganizations != nil {
			maxOrganizations = tenant.MaxOrganizations
		}
	}
	if maxOrganizations != nil {
		if maxOrganizations.Value == 0 {
			return errEntityQuotaReached.WithAttributes("entity_type", "organization")
		}
		count, err := countInTenant(org.Model.ctx, db, "organization")
		if err != nil {
			return err
		}
		if count >= maxOrganizations.Value {
			return errEntityQuotaReached.WithAttributes("entity_type", "organization")
		}
	}
	return nil
}

// BeforeCreate checks if the create is allowed by the license.
func (usr *User) BeforeCreate(db *gorm.DB) error {
	license, err := checkLicense(&usr.Model)
	if err != nil {
		return err
	}
	var maxUsers *types.UInt64Value
	if license.MaxUsers != nil {
		maxUsers = license.MaxUsers
	} else if license.MultiTenancy {
		tenant, err := retrieveTenant(db, &usr.Model)
		if err != nil {
			return err
		}
		if tenant.MaxUsers != nil {
			maxUsers = tenant.MaxUsers
		}
	}
	if maxUsers != nil {
		if maxUsers.Value == 0 {
			return errEntityQuotaReached.WithAttributes("entity_type", "user")
		}
		count, err := countInTenant(usr.Model.ctx, db, "user")
		if err != nil {
			return err
		}
		if count >= maxUsers.Value {
			return errEntityQuotaReached.WithAttributes("entity_type", "user")
		}
	}
	return nil
}
