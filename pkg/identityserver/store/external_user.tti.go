// Copyright © 2019 The Things Industries B.V.

package store

import "go.thethings.network/lorawan-stack/v3/pkg/ttipb"

// External user model.
type ExternalUser struct {
	Model
	SoftDelete

	TenantID string `gorm:"unique_index:uix_external_users_user_id;type:VARCHAR(36)"`

	User   *User
	UserID string `gorm:"type:UUID;index:uix_external_users_user_id;not null"`

	AuthenticationProvider   *AuthenticationProvider
	AuthenticationProviderID string `gorm:"type:UUID;not null"`

	ExternalID string `gorm:"not null"`
}

func (ExternalUser) _isMultiTenant() {}

func (eu ExternalUser) toPB() *ttipb.ExternalUser {
	pb := &ttipb.ExternalUser{
		CreatedAt:  cleanTime(eu.CreatedAt),
		UpdatedAt:  cleanTime(eu.UpdatedAt),
		ExternalID: eu.ExternalID,
	}
	if eu.User != nil {
		pb.UserIDs.UserID = eu.User.Account.UID
	}
	if eu.AuthenticationProvider != nil {
		pb.ProviderIDs.ProviderID = eu.AuthenticationProvider.ProviderID
	}
	return pb
}

func init() {
	registerModel(&ExternalUser{})
}
