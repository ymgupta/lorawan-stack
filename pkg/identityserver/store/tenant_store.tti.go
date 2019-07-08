// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"
	"fmt"
	"reflect"
	"runtime/trace"
	"strings"

	ptypes "github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/warning"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// GetTenantStore returns an TenantStore on the given db (or transaction).
func GetTenantStore(db *gorm.DB) TenantStore {
	return &tenantStore{store: newStore(db)}
}

type tenantStore struct {
	*store
}

// selectTenantFields selects relevant fields (based on fieldMask) and preloads details if needed.
func selectTenantFields(ctx context.Context, query *gorm.DB, fieldMask *ptypes.FieldMask) *gorm.DB {
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		return query.Preload("Attributes")
	}
	var tenantColumns []string
	var notFoundPaths []string
	for _, path := range ttnpb.TopLevelFields(fieldMask.Paths) {
		switch path {
		case "ids", "created_at", "updated_at":
			// always selected
		case attributesField:
			query = query.Preload("Attributes")
		default:
			if columns, ok := tenantColumnNames[path]; ok {
				tenantColumns = append(tenantColumns, columns...)
			} else {
				notFoundPaths = append(notFoundPaths, path)
			}
		}
	}
	if len(notFoundPaths) > 0 {
		warning.Add(ctx, fmt.Sprintf("unsupported field mask paths: %s", strings.Join(notFoundPaths, ", ")))
	}
	return query.Select(cleanFields(append(append(modelColumns, "tenant_id"), tenantColumns...)...))
}

func (s *tenantStore) CreateTenant(ctx context.Context, tnt *ttipb.Tenant) (*ttipb.Tenant, error) {
	defer trace.StartRegion(ctx, "create tenant").End()
	tntModel := Tenant{
		TenantID: tnt.TenantID, // The ID is not mutated by fromPB.
	}
	tntModel.fromPB(tnt, nil)
	if err := s.createEntity(ctx, &tntModel); err != nil {
		return nil, err
	}
	var tntProto ttipb.Tenant
	tntModel.toPB(&tntProto, nil)
	return &tntProto, nil
}

func (s *tenantStore) FindTenants(ctx context.Context, ids []*ttipb.TenantIdentifiers, fieldMask *ptypes.FieldMask) ([]*ttipb.Tenant, error) {
	defer trace.StartRegion(ctx, "find tenants").End()
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.GetTenantID()
	}
	query := s.query(ctx, Tenant{}, withTenantID(idStrings...))
	query = selectTenantFields(ctx, query, fieldMask)
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		countTotal(ctx, query.Model(&Tenant{}))
		query = query.Limit(limit).Offset(offset)
	}
	var tntModels []Tenant
	query = query.Find(&tntModels)
	setTotal(ctx, uint64(len(tntModels)))
	if query.Error != nil {
		return nil, query.Error
	}
	tntProtos := make([]*ttipb.Tenant, len(tntModels))
	for i, tntModel := range tntModels {
		tntProto := &ttipb.Tenant{}
		tntModel.toPB(tntProto, fieldMask)
		tntProtos[i] = tntProto
	}
	return tntProtos, nil
}

func (s *tenantStore) GetTenant(ctx context.Context, id *ttipb.TenantIdentifiers, fieldMask *ptypes.FieldMask) (*ttipb.Tenant, error) {
	defer trace.StartRegion(ctx, "get tenant").End()
	query := s.query(ctx, Tenant{}, withTenantID(id.GetTenantID()))
	query = selectTenantFields(ctx, query, fieldMask)
	var tntModel Tenant
	if err := query.First(&tntModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(id)
		}
		return nil, err
	}
	tntProto := &ttipb.Tenant{}
	tntModel.toPB(tntProto, fieldMask)
	return tntProto, nil
}

func (s *tenantStore) UpdateTenant(ctx context.Context, tnt *ttipb.Tenant, fieldMask *ptypes.FieldMask) (updated *ttipb.Tenant, err error) {
	defer trace.StartRegion(ctx, "update tenant").End()
	query := s.query(ctx, Tenant{}, withTenantID(tnt.GetTenantID()))
	query = selectTenantFields(ctx, query, fieldMask)
	var tntModel Tenant
	if err = query.First(&tntModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(tnt.TenantIdentifiers)
		}
		return nil, err
	}
	if err := ctx.Err(); err != nil { // Early exit if context canceled
		return nil, err
	}
	oldAttributes := tntModel.Attributes
	columns := tntModel.fromPB(tnt, fieldMask)
	if err = s.updateEntity(ctx, &tntModel, columns...); err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(oldAttributes, tntModel.Attributes) {
		if err = s.replaceAttributes(ctx, "tenant", tntModel.ID, oldAttributes, tntModel.Attributes); err != nil {
			return nil, err
		}
	}
	updated = &ttipb.Tenant{}
	tntModel.toPB(updated, fieldMask)
	return updated, nil
}

func (s *tenantStore) DeleteTenant(ctx context.Context, id *ttipb.TenantIdentifiers) error {
	defer trace.StartRegion(ctx, "delete tenant").End()
	return s.deleteEntity(ctx, id)
}

var errGatewayEUINotFound = errors.DefineNotFound("gateway_eui_not_found", "gateway eui `{eui}` not found")

func (s *tenantStore) GetTenantIDForGatewayEUI(ctx context.Context, eui types.EUI64) (*ttipb.TenantIdentifiers, error) {
	defer trace.StartRegion(ctx, "get tenant id for gateway eui").End()
	query := s.query(ctx, nil, withGatewayEUI(EUI64(eui))).Select("tenant_id")
	var gtwModel Gateway
	if err := query.First(&gtwModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errGatewayEUINotFound.WithAttributes("eui", eui.String())
		}
		return nil, err
	}
	return &ttipb.TenantIdentifiers{TenantID: gtwModel.TenantID}, nil
}
