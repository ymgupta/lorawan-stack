// Copyright © 2019 The Things Industries B.V.

package ttipb

import "go.thethings.network/lorawan-stack/pkg/ttnpb"

func onlyPublicContactInfo(info []*ttnpb.ContactInfo) []*ttnpb.ContactInfo {
	if info == nil {
		return nil
	}
	out := make([]*ttnpb.ContactInfo, 0, len(info))
	for _, info := range info {
		if !info.Public {
			continue
		}
		out = append(out, info)
	}
	return out
}

// PublicTenantFields are the Tenant's fields that are public.
var PublicTenantFields = append(ttnpb.PublicEntityFields,
	"name",
	"state",
	"capabilities",
)

// PublicSafe returns a copy of the tenant with only the fields that are safe to
// return to any audience.
func (t *Tenant) PublicSafe() *Tenant {
	if t == nil {
		return nil
	}
	var safe Tenant
	safe.SetFields(t, PublicTenantFields...)
	safe.ContactInfo = onlyPublicContactInfo(safe.ContactInfo)
	return &safe
}
