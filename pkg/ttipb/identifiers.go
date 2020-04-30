// Copyright © 2019 The Things Industries B.V.

package ttipb

// IsZero returns true if all identifiers have zero-values.
func (ids TenantIdentifiers) IsZero() bool {
	return ids.TenantID == ""
}
