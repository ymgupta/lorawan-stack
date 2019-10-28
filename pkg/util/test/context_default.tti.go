// Copyright © 2019 The Things Industries B.V.

// +build tti

package test

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func init() {
	DefaultContext = TenantContextFiller(context.Background())
}

func TenantContextFiller(ctx context.Context) context.Context {
	tenantID := tenant.FromContext(ctx)
	if tenantID.IsZero() {
		tenantID = ttipb.TenantIdentifiers{TenantID: "foo-tenant"}
		ctx = tenant.NewContext(ctx, tenantID)
	}
	ctx = tenant.NewContextWithFetcher(ctx, tenant.NewMapFetcher(map[string]*ttipb.Tenant{
		tenantID.TenantID: {TenantIdentifiers: tenantID, State: ttnpb.STATE_APPROVED},
	}))
	return ctx
}
