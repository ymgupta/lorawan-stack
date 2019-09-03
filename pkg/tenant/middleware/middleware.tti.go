// Copyright © 2019 The Things Industries B.V.

package middleware

import (
	"strings"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

var errMissingTenantID = errors.DefineInvalidArgument("missing_tenant_id", "missing tenant ID")

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
