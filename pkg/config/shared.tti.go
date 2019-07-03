// Copyright © 2019 The Things Industries B.V.

//+build tti

package config

// Tenancy represents configuration for tenancy.
type Tenancy struct {
	DefaultID string `name:"default-id" description:"Default tenant ID"`
}
