// Copyright © 2019 The Things Industries B.V.

package ttipb

import "go.thethings.network/lorawan-stack/pkg/ttnpb"

// EntityType returns the entity type for this ID (tenant).
func (TenantIdentifiers) EntityType() string { return "tenant" }

// IDString returns the ID string of this Identifier.
func (ids TenantIdentifiers) IDString() string { return ids.TenantID }

// Identifiers returns itself.
func (ids TenantIdentifiers) Identifiers() ttnpb.Identifiers { return &ids }

// EntityIdentifiers satisfies the ttnpb.Identifiers interface. It panic is actually used.
func (ids TenantIdentifiers) EntityIdentifiers() *ttnpb.EntityIdentifiers {
	panic("TenantIdentifiers aren't ttnpb.Identifiers")
}

// CombinedIdentifiers satisfies the ttnpb.Identifiers interface. It panic is actually used.
func (ids TenantIdentifiers) CombinedIdentifiers() *ttnpb.CombinedIdentifiers {
	panic("TenantIdentifiers aren't ttnpb.Identifiers")
}
