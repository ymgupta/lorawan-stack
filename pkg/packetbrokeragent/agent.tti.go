// Copyright © 2020 The Things Industries B.V.

package packetbrokeragent

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
)

// WithTenancyContextFiller returns an Option that fills the tenant context.
func WithTenancyContextFiller() Option {
	return WithTenantContextFiller(func(parent context.Context, tenantID string) (context.Context, error) {
		if tenantID == "" {
			tenantID = cluster.PacketBrokerTenantID.TenantID
		}
		return tenant.NewContext(parent, ttipb.TenantIdentifiers{TenantID: tenantID}), nil
	})
}
