// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"
	"fmt"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/tenant"
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
			if table == "users" || table == "organizations" {
				return db.Where("accounts.tenant_id = ?", tenant.FromContext(ctx).TenantID)
			}
			return db.Where(fmt.Sprintf("%s.tenant_id = ?", table), tenant.FromContext(ctx).TenantID)
		}
		return db
	}
}
