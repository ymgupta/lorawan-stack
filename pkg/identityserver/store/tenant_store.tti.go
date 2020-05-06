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
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/warning"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
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
	if _, err := tntModel.fromPB(tnt, nil); err != nil {
		return nil, err
	}
	if err := s.createEntity(ctx, &tntModel); err != nil {
		return nil, err
	}
	var tntProto ttipb.Tenant
	if err := tntModel.toPB(&tntProto, nil); err != nil {
		return nil, err
	}
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
	query = query.Order(orderFromContext(ctx, "tenants", "tenant_id", "ASC"))
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
		if err := tntModel.toPB(tntProto, fieldMask); err != nil {
			return nil, err
		}
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
	if err := tntModel.toPB(tntProto, fieldMask); err != nil {
		return nil, err
	}
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
	columns, err := tntModel.fromPB(tnt, fieldMask)
	if err != nil {
		return nil, err
	}
	if err = s.updateEntity(ctx, &tntModel, columns...); err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(oldAttributes, tntModel.Attributes) {
		if err = s.replaceAttributes(ctx, "tenant", tntModel.ID, oldAttributes, tntModel.Attributes); err != nil {
			return nil, err
		}
	}
	updated = &ttipb.Tenant{}
	if err = tntModel.toPB(updated, fieldMask); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *tenantStore) DeleteTenant(ctx context.Context, id *ttipb.TenantIdentifiers) error {
	defer trace.StartRegion(ctx, "delete tenant").End()
	return s.deleteEntity(ctx, id)
}

var errEndDeviceEUIsNotFound = errors.DefineNotFound("end_device_euis_not_found", "end device JoinEUI `{join_eui}` and DevEUI `{dev_eui}` not found")

func (s *tenantStore) GetTenantIDForEndDeviceEUIs(ctx context.Context, joinEUI, devEUI types.EUI64) (*ttipb.TenantIdentifiers, error) {
	defer trace.StartRegion(ctx, "get tenant id for end device euis").End()
	query := s.query(ctx, nil, withJoinEUI(EUI64(joinEUI)), withDevEUI(EUI64(devEUI))).Select("tenant_id")
	var devModel EndDevice
	if err := query.First(&devModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errEndDeviceEUIsNotFound.WithAttributes("join_eui", joinEUI.String(), "dev_eui", devEUI.String())
		}
		return nil, err
	}
	return &ttipb.TenantIdentifiers{TenantID: devModel.TenantID}, nil
}

var errGatewayEUINotFound = errors.DefineNotFound("gateway_eui_not_found", "gateway EUI `{eui}` not found")

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

func (s *tenantStore) CountEntities(ctx context.Context, id *ttipb.TenantIdentifiers, entityType string) (uint64, error) {
	defer trace.StartRegion(ctx, "count entities for tenant id").End()
	var total uint64
	query := s.query(ctx, nil)
	if id != nil {
		query = withTenantID(id.TenantID)(query)
	}
	switch entityType {
	case "organization", "user":
		query = query.Model(Account{}).Where(Account{AccountType: entityType})
	default:
		query = query.Model(modelForEntityType(entityType))
	}
	err := query.Count(&total).Error
	return total, err
}
