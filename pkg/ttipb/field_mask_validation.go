// Copyright © 2019 The Things Industries B.V.

package ttipb

// AllowedFieldMaskPathsForRPC lists the allowed field mask paths for each RPC in this API.
var AllowedFieldMaskPathsForRPC = map[string][]string{
	// Tenants:
	"/tti.lorawan.v3.TenantRegistry/Get":               TenantFieldPathsNested,
	"/tti.lorawan.v3.TenantRegistry/GetRegistryTotals": TenantRegistryTotalsFieldPathsNested,
	"/tti.lorawan.v3.TenantRegistry/List":              TenantFieldPathsNested,
	"/tti.lorawan.v3.TenantRegistry/Update":            TenantFieldPathsNested,
}
