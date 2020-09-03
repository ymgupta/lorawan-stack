// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"
	"fmt"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/license"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
)

func init() {
	contextScoper = func(ctx context.Context, db *gorm.DB) *gorm.DB {
		if db.Value == nil {
			return db
		}
		if !db.NewScope(db.Value).PrimaryKeyZero() {
			return db
		}
		if _, ok := db.Value.(interface{ _isMultiTenant() }); ok {
			table := db.NewScope(db.Value).TableName()
			tenantID := tenant.FromContext(ctx).TenantID
			if table == "users" || table == "organizations" {
				return db.Where("accounts.tenant_id = ?", tenantID)
			}
			if table == "clients" && tenantID == "" && license.RequireMultiTenancy(ctx) == nil {
				return db.Where(fmt.Sprintf("%s.tenant_id IS NULL", table))
			}
			return db.Where(fmt.Sprintf("%s.tenant_id = ?", table), tenantID)
		}
		return db
	}
}

func withTenantID(id ...string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		switch len(id) {
		case 0:
			return db
		case 1:
			return db.Where("tenant_id = ?", id[0])
		default:
			return db.Where("tenant_id IN (?)", id).Order("tenant_id")
		}
	}
}

func withProviderID(id ...string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		switch len(id) {
		case 0:
			return db
		case 1:
			return db.Where("provider_id = ?", id[0])
		default:
			return db.Where("provider_id IN (?)", id).Order("provider_id")
		}
	}
}
