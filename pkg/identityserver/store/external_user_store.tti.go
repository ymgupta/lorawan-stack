// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"
	"runtime/trace"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// GetExternalUserStore returns an ExternalUserStore on the given db (or transaction).
func GetExternalUserStore(db *gorm.DB) ExternalUserStore {
	return &externalUserStore{store: newStore(db)}
}

type externalUserStore struct {
	*store
}

func (s *externalUserStore) CreateExternalUser(ctx context.Context, eu *ttipb.ExternalUser) (*ttipb.ExternalUser, error) {
	defer trace.StartRegion(ctx, "create external user").End()
	user, err := s.findEntity(ctx, eu.UserIDs, "id")
	if err != nil {
		return nil, err
	}
	provider := &AuthenticationProvider{}
	err = s.query(ctx, AuthenticationProvider{}, withProviderID(eu.ProviderIDs.ProviderID)).
		Select("id").
		First(provider).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errAuthenticationProviderNotFound.WithAttributes("provider_id", eu.ProviderIDs.ProviderID)
		}
		return nil, err
	}
	model := ExternalUser{
		UserID:                   user.PrimaryKey(),
		AuthenticationProviderID: provider.ID,
		ExternalID:               eu.ExternalID,
	}
	model.CreatedAt = cleanTime(eu.CreatedAt)
	if err := s.createEntity(ctx, &model); err != nil {
		return nil, err
	}
	proto := model.toPB()
	proto.UserIDs = eu.UserIDs
	proto.ProviderIDs = eu.ProviderIDs
	return proto, nil
}

type friendlyExternalUser struct {
	ExternalUser
	FriendlyUserID string
}

func (s *externalUserStore) optimizedQuery(ctx context.Context) *gorm.DB {
	return s.store.query(ctx, ExternalUser{}).
		Select(`"external_users".*, "accounts"."uid" AS "friendly_user_id"`).
		Joins(`LEFT JOIN "authentication_providers" ON "authentication_providers"."id" = "external_users"."authentication_provider_id"`).
		Joins(`LEFT JOIN "users" ON "users"."id" = "external_users"."user_id"`).
		Joins(`LEFT JOIN "accounts" ON "accounts"."account_type" = 'user' AND "accounts"."account_id" = "users"."id"`)
}

func (s *externalUserStore) GetExternalUserByUserID(ctx context.Context, ids *ttnpb.UserIdentifiers) (*ttipb.ExternalUser, error) {
	defer trace.StartRegion(ctx, "get external user by user id").End()
	var model friendlyExternalUser
	err := s.optimizedQuery(ctx).
		Where(Account{UID: ids.GetUserID()}).
		Scan(&model).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errExternalUserNotFound.WithAttributes("user_id", ids.GetUserID())
		}
		return nil, err
	}
	proto := model.toPB()
	proto.UserIDs.UserID = model.FriendlyUserID
	return proto, nil
}

func (s *externalUserStore) GetExternalUserByExternalID(ctx context.Context, providerIDs *ttipb.AuthenticationProviderIdentifiers, externalID string) (*ttipb.ExternalUser, error) {
	defer trace.StartRegion(ctx, "get external user by user id").End()
	var model friendlyExternalUser
	err := s.optimizedQuery(ctx).
		Where(ExternalUser{ExternalID: externalID}).
		Where(AuthenticationProvider{ProviderID: providerIDs.ProviderID}).
		Scan(&model).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errExternalUserNotFound.WithAttributes("user_id", externalID)
		}
		return nil, err
	}
	proto := model.toPB()
	proto.UserIDs.UserID = model.FriendlyUserID
	return proto, nil
}

func (s *externalUserStore) DeleteExternalUser(ctx context.Context, providerIDs *ttipb.AuthenticationProviderIdentifiers, externalID string) error {
	defer trace.StartRegion(ctx, "delete external user").End()
	var model friendlyExternalUser
	err := s.optimizedQuery(ctx).
		Where(ExternalUser{ExternalID: externalID}).
		Where(AuthenticationProvider{ProviderID: providerIDs.ProviderID}).
		Scan(&model).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errExternalUserNotFound.WithAttributes("user_id", externalID)
		}
		return err
	}
	return s.DB.Delete(model.ExternalUser).Error
}
