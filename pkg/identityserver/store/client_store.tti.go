// Copyright © 2019 The Things Industries B.V.

package store

import (
	"context"
	"net/url"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/license"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func (s *clientStore) getClientWithoutTenant(ctx context.Context, id *ttnpb.ClientIdentifiers, fieldMask *types.FieldMask) (*ttnpb.Client, error) {
	tenantID := tenant.FromContext(ctx).TenantID
	if license.RequireMultiTenancy(ctx) != nil || tenantID == "" {
		return nil, errNotFoundForID(id)
	}
	query := s.query(tenant.NewContext(ctx, ttipb.TenantIdentifiers{}), Client{}, withClientID(id.GetClientID()))
	query = selectClientFields(ctx, query, fieldMask)
	var cliModel Client
	if err := query.First(&cliModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errNotFoundForID(id)
		}
		return nil, err
	}
	cliProto := &ttnpb.Client{}
	cliModel.toPB(cliProto, fieldMask)

	// Add tenant ID as prefix in Redirect URIs:
	if fieldPaths := fieldMask.GetPaths(); len(fieldPaths) == 0 || ttnpb.HasAnyField(fieldPaths, "redirect_uris") {
		var tenantRedirectURIs []string
		for _, redirectURI := range cliProto.RedirectURIs {
			if !strings.Contains(redirectURI, "://") {
				continue
			}
			if uri, err := url.Parse(redirectURI); err == nil {
				uri.Host = tenantID + "." + uri.Host
				tenantRedirectURIs = append(tenantRedirectURIs, uri.String())
			}
		}
		if len(tenantRedirectURIs) > 0 {
			cliProto.RedirectURIs = append(cliProto.RedirectURIs, tenantRedirectURIs...)
		}
	}

	// Add tenant ID as prefix in Logout Redirect URIs:
	if fieldPaths := fieldMask.GetPaths(); len(fieldPaths) == 0 || ttnpb.HasAnyField(fieldPaths, "logout_redirect_uris") {
		var tenantLogoutRedirectURIs []string
		for _, logoutRedirectURI := range cliProto.LogoutRedirectURIs {
			if !strings.Contains(logoutRedirectURI, "://") {
				continue
			}
			if uri, err := url.Parse(logoutRedirectURI); err == nil {
				uri.Host = tenantID + "." + uri.Host
				tenantLogoutRedirectURIs = append(tenantLogoutRedirectURIs, uri.String())
			}
		}
		if len(tenantLogoutRedirectURIs) > 0 {
			cliProto.LogoutRedirectURIs = append(cliProto.LogoutRedirectURIs, tenantLogoutRedirectURIs...)
		}
	}

	return cliProto, nil
}
