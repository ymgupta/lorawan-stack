// Copyright © 2019 The Things Industries B.V.

//+build tti

package tenant

import "go.thethings.network/lorawan-stack/pkg/errors"

// AllowEmptyTenantID makes the tenant package allow tenant IDs to be missing.
// This is useful in case a default tenant ID of "" is used.
func AllowEmptyTenantID() {
	allowEmptyTenantID = true
}

var allowEmptyTenantID bool

var errMissingTenantID = errors.DefineInvalidArgument("missing_tenant_id", "missing tenant ID")

// UseEmptyID can be called to indicate that an empty Tenant ID is used. It returns
// an error if that is not acceptable.
// This function is typically called an API boundary or by middleware.
func UseEmptyID() error {
	if !allowEmptyTenantID {
		return errMissingTenantID
	}
	return nil
}
