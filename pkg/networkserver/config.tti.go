// Copyright © 2020 The Things Industries B.V.

package networkserver

// EnterpriseConfig represents enterprise-specific configuration.
type EnterpriseConfig struct {
	SwitchPeeringTenantContext bool `name:"switch-peering-tenant-context" description:"Set to 'true' to switch tenant context in peering"`
}
