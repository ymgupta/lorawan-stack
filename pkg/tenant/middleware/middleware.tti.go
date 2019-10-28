// Copyright © 2019 The Things Industries B.V.

package middleware

import (
	"context"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	errMissingTenantID = errors.DefineInvalidArgument("missing_tenant_id", "missing tenant ID")
	errTenantNotActive = errors.DefinePermissionDenied("tenant_not_active", "tenant is not active")
)

// tenantID parses the tenant ID from the given value.
// If the given value contains dots (i.e. a host name), the first part is assumed to be the tenant ID.
func tenantID(v string) string {
	if idx := strings.Index(v, "."); idx != -1 {
		v = v[:idx]
	}
	id := ttipb.TenantIdentifiers{TenantID: v}
	if id.ValidateFields("tenant_id") != nil {
		return ""
	}
	return id.TenantID
}

func fetchTenant(ctx context.Context) error {
	if tenantFetcher, ok := tenant.FetcherFromContext(ctx); ok {
		tenantID := tenant.FromContext(ctx)
		tnt, err := tenantFetcher.FetchTenant(ctx, &tenantID, "name", "state")
		if err != nil {
			return err
		}
		switch tnt.State {
		case ttnpb.STATE_REQUESTED, ttnpb.STATE_REJECTED, ttnpb.STATE_SUSPENDED:
			return errTenantNotActive.WithAttributes("state", tnt.State)
		case ttnpb.STATE_APPROVED, ttnpb.STATE_FLAGGED:
			break
		default:
			panic("unreachable")
		}
	}
	return nil
}
