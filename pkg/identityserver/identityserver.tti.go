// Copyright © 2019 The Things Industries B.V.

package identityserver

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/oauth"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
)

func tenantConfigFromContext(ctx context.Context) (*ttipb.Configuration_Cluster_IdentityServer, bool) {
	conf, err := tenant.ConfigFromContext(ctx)
	if err != nil {
		return nil, false
	}
	isConf := conf.GetDefaultCluster().GetIS()
	return isConf, isConf != nil
}

func (conf Config) Apply(ctx context.Context) Config {
	deriv := conf
	deriv.Email.Config = conf.Email.Config.Apply(ctx)

	tenantConf, ok := tenantConfigFromContext(ctx)
	if !ok {
		return deriv
	}
	userRegistration := tenantConf.GetUserRegistration()
	if required := userRegistration.GetInvitation().GetRequired(); required != nil {
		deriv.UserRegistration.Invitation.Required = required.Value
	}
	if tokenTTL := userRegistration.GetInvitation().GetTokenTTL(); tokenTTL != nil {
		deriv.UserRegistration.Invitation.TokenTTL = *tokenTTL
	}
	if required := userRegistration.GetContactInfoValidation().GetRequired(); required != nil {
		deriv.UserRegistration.ContactInfoValidation.Required = required.Value
	}
	if required := userRegistration.GetAdminApproval().GetRequired(); required != nil {
		deriv.UserRegistration.AdminApproval.Required = required.Value
	}
	if w := userRegistration.GetPasswordRequirements().GetMinLength(); w != nil {
		deriv.UserRegistration.PasswordRequirements.MinLength = int(w.Value)
	}
	if w := userRegistration.GetPasswordRequirements().GetMaxLength(); w != nil {
		deriv.UserRegistration.PasswordRequirements.MaxLength = int(w.Value)
	}
	if w := userRegistration.GetPasswordRequirements().GetMinUppercase(); w != nil {
		deriv.UserRegistration.PasswordRequirements.MinUppercase = int(w.Value)
	}
	if w := userRegistration.GetPasswordRequirements().GetMinDigits(); w != nil {
		deriv.UserRegistration.PasswordRequirements.MinDigits = int(w.Value)
	}
	if w := userRegistration.GetPasswordRequirements().GetMinSpecial(); w != nil {
		deriv.UserRegistration.PasswordRequirements.MinSpecial = int(w.Value)
	}
	profilePicture := tenantConf.GetProfilePicture()
	if w := profilePicture.GetDisableUpload(); w != nil {
		deriv.ProfilePicture.DisableUpload = w.Value
	}
	if w := profilePicture.GetUseGravatar(); w != nil {
		deriv.ProfilePicture.UseGravatar = w.Value
	}
	endDevicePicture := tenantConf.GetEndDevicePicture()
	if w := endDevicePicture.GetDisableUpload(); w != nil {
		deriv.EndDevicePicture.DisableUpload = w.Value
	}
	userRights := tenantConf.GetUserRights()
	if w := userRights.GetCreateApplications(); w != nil {
		deriv.UserRights.CreateApplications = w.Value
	}
	if w := userRights.GetCreateClients(); w != nil {
		deriv.UserRights.CreateClients = w.Value
	}
	if w := userRights.GetCreateGateways(); w != nil {
		deriv.UserRights.CreateGateways = w.Value
	}
	if w := userRights.GetCreateOrganizations(); w != nil {
		deriv.UserRights.CreateOrganizations = w.Value
	}
	return deriv
}

// TenancyConfig is the configuration for tenancy in the Identity Server.
type TenancyConfig struct {
	AdminKeys []string `name:"admin-keys" description:"Keys that can be used for tenant administration"`

	decodedAdminKeys [][]byte
}

func (c *TenancyConfig) decodeAdminKeys(ctx context.Context) error {
	for i, key := range c.AdminKeys {
		decodedKey, err := hex.DecodeString(key)
		if err != nil {
			return errInvalidTenantAdminKey.WithCause(err)
		}
		switch len(decodedKey) {
		case 16, 24, 32:
		default:
			return errInvalidTenantAdminKey.WithCause(fmt.Errorf("invalid length for key %d: must be 16, 24 or 32 bytes", i))
		}
		c.decodedAdminKeys = append(c.decodedAdminKeys, decodedKey)
	}
	if c.decodedAdminKeys == nil {
		c.decodedAdminKeys = [][]byte{random.Bytes(32)}
		log.FromContext(ctx).WithField("key", hex.EncodeToString(c.decodedAdminKeys[0])).Warn("No tenant admin key configured, generated a random one")
	}
	return nil
}

func (is *IdentityServer) withReadDatabase(ctx context.Context, f func(*gorm.DB) error) error {
	db := is.db
	if is.readDB != nil {
		db = is.readDB
	}
	return store.ReadOnly(ctx, db, f)
}

func withOAuthConfigPatcher(ctx context.Context) context.Context {
	return oauth.WithConfigurationPatcher(ctx, oauth.ConfigurationPatcherFunc(
		func(ctx context.Context, conf oauth.Config) oauth.Config {
			deriv := conf
			isConf, ok := tenantConfigFromContext(ctx)
			if !ok {
				return deriv
			}
			oauth := isConf.GetOAuth()
			providers := oauth.GetProviders()
			if oidc := providers.GetOIDC(); oidc != nil {
				deriv.Providers.OIDC.Name = oidc.Name
				deriv.Providers.OIDC.AllowRegistrations = oidc.AllowRegistrations
				deriv.Providers.OIDC.ClientID = oidc.ClientID
				deriv.Providers.OIDC.ClientSecret = oidc.ClientSecret
				deriv.Providers.OIDC.RedirectURL = oidc.RedirectURL
				deriv.Providers.OIDC.ProviderURL = oidc.ProviderURL
			}
			return deriv
		},
	))
}
