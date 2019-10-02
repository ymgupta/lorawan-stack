// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/license"
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
