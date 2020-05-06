// Copyright © 2019 The Things Industries B.V.

package middleware

import (
	"context"
	"fmt"
	"net"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	errMissingTenantID = errors.DefineInvalidArgument("missing_tenant_id", "missing tenant ID")
	errTenantNotActive = errors.DefineFailedPrecondition("tenant_not_active", "tenant is not active", "state")
)

// tenantID parses the tenant ID from the given value.
// If the given value contains dots (i.e. a host name), the first part is assumed to be the tenant ID.
func tenantID(v string, config tenant.Config) string {
	if len(config.BaseDomains) > 0 {
		if host, _, err := net.SplitHostPort(v); err == nil {
			v = host
		}
		// Strip base domain. Also works if v equals a base domain.
		for _, suffix := range config.BaseDomains {
			if strings.HasSuffix(v, suffix) {
				v = strings.TrimSuffix(v, suffix)
				break
			}
		}
		if strings.HasSuffix(v, ".") {
			v = strings.TrimSuffix(v, ".")
		}
	} else {
		// Old behavior (before base domains were added to config).
		if idx := strings.Index(v, "."); idx != -1 {
			v = v[:idx]
		}
	}
	if v == "" {
		return ""
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
			panic(fmt.Sprintf("Unhandled tenant state: %s", tnt.State.String()))
		}
	}
	return nil
}
