// Copyright © 2019 The Things Industries B.V.

// Package tenant contains context handling and middleware for tenancy.
package tenant

import "go.thethings.network/lorawan-stack/pkg/errors"

// Config represents tenancy configuration.
type Config struct {
	DefaultID string `name:"default-id" description:"Default tenant ID"`
}

var errMissingTenantID = errors.DefineInvalidArgument("missing_tenant_id", "missing tenant ID")
