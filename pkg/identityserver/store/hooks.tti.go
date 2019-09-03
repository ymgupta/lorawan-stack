// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

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
	if license.MaxApplications != nil {
		if license.MaxApplications.Value == 0 {
			return errEntityQuotaReached.WithAttributes("entity_type", "application")
		}
		count, err := countInTenant(app.Model.ctx, db, "application")
		if err != nil {
			return err
		}
		if count >= license.MaxApplications.Value {
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
	if license.MaxClients != nil {
		if license.MaxClients.Value == 0 {
			return errEntityQuotaReached.WithAttributes("entity_type", "client")
		}
		count, err := countInTenant(cli.Model.ctx, db, "client")
		if err != nil {
			return err
		}
		if count >= license.MaxClients.Value {
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
	if license.MaxEndDevices != nil {
		if license.MaxEndDevices.Value == 0 {
			return errEntityQuotaReached.WithAttributes("entity_type", "end device")
		}
		count, err := countInTenant(dev.Model.ctx, db, "end_device")
		if err != nil {
			return err
		}
		if count >= license.MaxEndDevices.Value {
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
	if license.MaxGateways != nil {
		if license.MaxGateways.Value == 0 {
			return errEntityQuotaReached.WithAttributes("entity_type", "gateway")
		}
		count, err := countInTenant(gtw.Model.ctx, db, "gateway")
		if err != nil {
			return err
		}
		if count >= license.MaxGateways.Value {
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
	if license.MaxOrganizations != nil {
		if license.MaxOrganizations.Value == 0 {
			return errEntityQuotaReached.WithAttributes("entity_type", "organization")
		}
		count, err := countInTenant(org.Model.ctx, db, "organization")
		if err != nil {
			return err
		}
		if count >= license.MaxOrganizations.Value {
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
	if license.MaxUsers != nil {
		if license.MaxUsers.Value == 0 {
			return errEntityQuotaReached.WithAttributes("entity_type", "user")
		}
		count, err := countInTenant(usr.Model.ctx, db, "user")
		if err != nil {
			return err
		}
		if count >= license.MaxUsers.Value {
			return errEntityQuotaReached.WithAttributes("entity_type", "user")
		}
	}
	return nil
}
