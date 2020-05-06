// Copyright © 2019 The Things Industries B.V.

package identityserver

import (
	"context"
	"crypto/subtle"
	"encoding/hex"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"google.golang.org/grpc"
)

var (
	// TenantAdminAuthType is the AuthType used for tenant administration.
	TenantAdminAuthType = "TenantAdminKey"

	errInvalidTenantAdminKey    = errors.DefinePermissionDenied("tenant_admin_key", "invalid tenant admin key")
	errNoTenantRights           = errors.DefinePermissionDenied("no_tenant_rights", "no tenant rights")
	errInsufficientTenantRights = errors.DefinePermissionDenied("insufficient_tenant_rights", "insufficient tenant rights")
)

type tenantRightsKeyType struct{}

var tenantRightsKey tenantRightsKeyType

type tenantRights struct {
	admin        bool
	read         bool
	writeCurrent bool
	readCurrent  bool
}

func (r tenantRights) canRead(ctx context.Context, tenantID *ttipb.TenantIdentifiers) bool {
	if r.admin || r.read {
		return true
	}
	if r.readCurrent && tenantID != nil && tenantID.Equal(tenant.FromContext(ctx)) {
		return true
	}
	return false
}

func (r tenantRights) canWrite(ctx context.Context, tenantID *ttipb.TenantIdentifiers) bool {
	if r.admin {
		return true
	}
	if r.writeCurrent && tenantID != nil && tenantID.Equal(tenant.FromContext(ctx)) {
		return true
	}
	return false
}

func tenantRightsFromContext(ctx context.Context) tenantRights {
	if rights, ok := ctx.Value(tenantRightsKey).(tenantRights); ok {
		return rights
	}
	return tenantRights{}
}

func newContextWithTenantRights(parent context.Context, rights tenantRights) context.Context {
	return context.WithValue(parent, tenantRightsKey, rights)
}

func (is *IdentityServer) tenantRightsHook(h grpc.UnaryHandler) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		md := rpcmetadata.FromIncomingContext(ctx)
		if md.AuthType == "" {
			return nil, errUnauthenticated.New()
		}
		rights := tenantRights{}
		switch strings.ToLower(md.AuthType) {
		case strings.ToLower(TenantAdminAuthType):
			key, err := hex.DecodeString(md.AuthValue)
			if err != nil {
				return nil, errInvalidTenantAdminKey.WithCause(err)
			}
			var isValidKey bool
			for _, acceptedKey := range is.config.Tenancy.decodedAdminKeys {
				if subtle.ConstantTimeCompare(acceptedKey, key) == 1 {
					isValidKey = true
				}
			}
			if !isValidKey {
				return nil, errInvalidTenantAdminKey.New()
			}
			rights.admin, rights.read = true, true
			rights.writeCurrent, rights.readCurrent = true, true
		case strings.ToLower(cluster.AuthType):
			if cluster.Authorized(ctx) == nil {
				rights.read, rights.readCurrent = true, true
			}
		case "bearer":
			tokenType, _, _, err := auth.SplitToken(md.AuthValue)
			if err != nil {
				return nil, err
			}
			switch tokenType {
			case auth.APIKey, auth.AccessToken:
				authInfo, err := is.authInfo(ctx)
				if err != nil {
					return nil, err
				}
				if authInfo.IsAdmin {
					rights.readCurrent, rights.writeCurrent = true, true
				}
			default:
				return nil, errUnsupportedAuthorization.New()
			}
		default:
			return nil, errUnsupportedAuthorization.New()
		}
		return h(newContextWithTenantRights(ctx, rights), req)
	}
}
