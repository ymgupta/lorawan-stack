// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"runtime/debug"
	"runtime/trace"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func (s *store) findEntityWithoutTenant(ctx context.Context, entityID ttnpb.Identifiers, fields ...string) (modelInterface, error) {
	tenantID := tenant.FromContext(ctx).TenantID
	if license.RequireMultiTenancy(ctx) != nil || tenantID == "" {
		return nil, errNotFoundForID(entityID)
	}

	var model modelInterface
	switch entityID.EntityType() {
	case "client":
		model = &Client{}
	default:
		return nil, errNotFoundForID(entityID)
	}

	query := s.query(tenant.NewContext(ctx, ttipb.TenantIdentifiers{}), model, withID(entityID))
	if len(fields) == 1 && fields[0] == "id" {
		fields[0] = s.DB.NewScope(model).TableName() + ".id"
	}
	if len(fields) > 0 {
		query = query.Select(fields)
	}
	if err := query.First(model).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(entityID)
		}
		return nil, convertError(err)
	}
	return model, nil
}

// ReadOnly is Transact, but then for read-only databases.
func ReadOnly(ctx context.Context, db *gorm.DB, f func(db *gorm.DB) error) (err error) {
	defer trace.StartRegion(ctx, "database transaction").End()
	tx := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if tx.Error != nil {
		return convertError(tx.Error)
	}
	defer func() {
		if p := recover(); p != nil {
			fmt.Fprintln(os.Stderr, p)
			os.Stderr.Write(debug.Stack())
			if pErr, ok := p.(error); ok {
				switch pErr {
				case context.Canceled, context.DeadlineExceeded:
					err = pErr
				default:
					err = ErrTransactionRecovered.WithCause(pErr)
				}
			} else {
				err = ErrTransactionRecovered.WithAttributes("panic", p)
			}
		}
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit().Error
		}
		err = convertError(err)
	}()
	SetLogger(tx, log.FromContext(ctx).WithField("namespace", "db"))
	return f(tx)
}
