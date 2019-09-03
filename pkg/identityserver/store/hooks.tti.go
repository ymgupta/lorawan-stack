// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/license"
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

func countInTenant(ctx context.Context, db *gorm.DB, model interface{}) (uint64, error) {
	var count uint64
	err := (&store{db}).query(ctx, model).Count(&count).Error
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
		count, err := countInTenant(app.Model.ctx, db, Application{})
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
		count, err := countInTenant(cli.Model.ctx, db, Client{})
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
		count, err := countInTenant(dev.Model.ctx, db, EndDevice{})
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
		count, err := countInTenant(gtw.Model.ctx, db, Gateway{})
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
		count, err := countInTenant(org.Model.ctx, db, Organization{})
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
		count, err := countInTenant(usr.Model.ctx, db, User{})
		if err != nil {
			return err
		}
		if count >= license.MaxUsers.Value {
			return errEntityQuotaReached.WithAttributes("entity_type", "user")
		}
	}
	return nil
}
