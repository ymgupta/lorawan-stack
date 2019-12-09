// Copyright © 2019 The Things Industries B.V.

// Package tenant contains context handling and middleware for tenancy.
package tenant

import "time"

// Config represents tenancy configuration.
type Config struct {
	DefaultID   string        `name:"default-id" description:"Default tenant ID"`
	BaseDomains []string      `name:"base-domains" description:"Base domains for tenant ID inference"`
	CacheTTL    time.Duration `name:"ttl" description:"TTL of cached tenant configurations"`
}
