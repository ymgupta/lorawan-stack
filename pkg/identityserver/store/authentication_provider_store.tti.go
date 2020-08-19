// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"
	"fmt"
	"runtime/trace"
	"strings"

	ptypes "github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/warning"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// GetAuthenticationProviderStore returns an AuthenticationProviderStore on the given db (or transaction).
func GetAuthenticationProviderStore(db *gorm.DB) AuthenticationProviderStore {
	return &authenticationProviderStore{store: newStore(db)}
}

type authenticationProviderStore struct {
	*store
}

// selectAuthenticationProviderFields selects relevant fields (based on fieldMask) and preloads details if needed.
func selectAuthenticationProviderFields(ctx context.Context, query *gorm.DB, fieldMask *ptypes.FieldMask) *gorm.DB {
	if fieldMask == nil || len(fieldMask.Paths) == 0 {
		return query
	}
	var providerColumns []string
	var notFoundPaths []string
	for _, path := range ttnpb.TopLevelFields(fieldMask.Paths) {
		switch path {
		case "ids", "created_at", "updated_at":
			// always selected
		default:
			if columns, ok := authenticationProviderColumnNames[path]; ok {
				providerColumns = append(providerColumns, columns...)
			} else {
				notFoundPaths = append(notFoundPaths, path)
			}
		}
	}
	if len(notFoundPaths) > 0 {
		warning.Add(ctx, fmt.Sprintf("unsupported field mask paths: %s", strings.Join(notFoundPaths, ", ")))
	}
	return query.Select(cleanFields(append(append(modelColumns, "provider_id"), providerColumns...)...))
}

func (s *authenticationProviderStore) CreateAuthenticationProvider(ctx context.Context, ap *ttipb.AuthenticationProvider) (*ttipb.AuthenticationProvider, error) {
	defer trace.StartRegion(ctx, "create authentication provider").End()
	model := AuthenticationProvider{
		ProviderID: ap.ProviderID,
	}
	if _, err := model.fromPB(ap, nil); err != nil {
		return nil, err
	}
	if err := s.createEntity(ctx, &model); err != nil {
		return nil, err
	}
	proto := &ttipb.AuthenticationProvider{}
	if err := model.toPB(proto, nil); err != nil {
		return nil, err
	}
	return proto, nil
}

func (s *authenticationProviderStore) FindAuthenticationProviders(ctx context.Context, ids []*ttipb.AuthenticationProviderIdentifiers, fieldMask *ptypes.FieldMask) ([]*ttipb.AuthenticationProvider, error) {
	defer trace.StartRegion(ctx, "find authentication providers").End()
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.GetProviderID()
	}
	query := s.query(ctx, AuthenticationProvider{}, withProviderID(idStrings...))
	query = selectAuthenticationProviderFields(ctx, query, fieldMask)
	query = query.Order(orderFromContext(ctx, "authentication_providers", "provider_id", "ASC"))
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		countTotal(ctx, query.Model(&AuthenticationProvider{}))
		query = query.Limit(limit).Offset(offset)
	}
	var models []AuthenticationProvider
	query = query.Find(&models)
	setTotal(ctx, uint64(len(models)))
	if query.Error != nil {
		return nil, query.Error
	}
	protos := make([]*ttipb.AuthenticationProvider, len(models))
	for i, model := range models {
		proto := &ttipb.AuthenticationProvider{}
		if err := model.toPB(proto, fieldMask); err != nil {
			return nil, err
		}
		protos[i] = proto
	}
	return protos, nil
}

func (s *authenticationProviderStore) GetAuthenticationProvider(ctx context.Context, ids *ttipb.AuthenticationProviderIdentifiers, fieldMask *ptypes.FieldMask) (*ttipb.AuthenticationProvider, error) {
	defer trace.StartRegion(ctx, "get authentication provider").End()
	query := s.query(ctx, AuthenticationProvider{}, withProviderID(ids.GetProviderID()))
	query = selectAuthenticationProviderFields(ctx, query, fieldMask)
	var model AuthenticationProvider
	if err := query.First(&model).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errAuthenticationProviderNotFound.WithAttributes("provider_id", ids.GetProviderID())
		}
		return nil, err
	}
	proto := &ttipb.AuthenticationProvider{}
	if err := model.toPB(proto, fieldMask); err != nil {
		return nil, err
	}
	return proto, nil
}

func (s *authenticationProviderStore) UpdateAuthenticationprovider(ctx context.Context, ap *ttipb.AuthenticationProvider, fieldMask *ptypes.FieldMask) (*ttipb.AuthenticationProvider, error) {
	defer trace.StartRegion(ctx, "update authentication provider").End()
	query := s.query(ctx, AuthenticationProvider{}, withProviderID(ap.GetProviderID()))
	query = selectAuthenticationProviderFields(ctx, query, fieldMask)
	var model AuthenticationProvider
	if err := query.First(&model).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errAuthenticationProviderNotFound.WithAttributes("provider_id", ap.GetProviderID())
		}
		return nil, err
	}
	if err := ctx.Err(); err != nil { // Early exit if context canceled
		return nil, err
	}
	columns, err := model.fromPB(ap, fieldMask)
	if err != nil {
		return nil, err
	}
	if err = s.updateEntity(ctx, &model, columns...); err != nil {
		return nil, err
	}
	updated := &ttipb.AuthenticationProvider{}
	if err = model.toPB(updated, fieldMask); err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *authenticationProviderStore) DeleteAuthenticationProvider(ctx context.Context, ids *ttipb.AuthenticationProviderIdentifiers) error {
	defer trace.StartRegion(ctx, "delete authentication provider").End()
	err := s.query(ctx, AuthenticationProvider{}).Where(AuthenticationProvider{ProviderID: ids.GetProviderID()}).Delete(AuthenticationProvider{}).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errAuthenticationProviderNotFound.WithAttributes("provider_id", ids.GetProviderID())
		}
		return err
	}
	return nil
}
