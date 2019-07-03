// Copyright © 2019 The Things Industries B.V.

//+build tti

package gatewayserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"google.golang.org/grpc"
)

// WithTenantRegistry overrides the CUPS server's tenant registry.
func WithTenantRegistry(registry ttipb.TenantRegistryClient) Option {
	return func(s *GatewayServer) {
		s.tenantRegistry = registry
	}
}

func (gs *GatewayServer) getTenantRegistry(ctx context.Context, ids *ttnpb.GatewayIdentifiers) ttipb.TenantRegistryClient {
	if gs.tenantRegistry != nil {
		return gs.tenantRegistry
	}
	return ttipb.NewTenantRegistryClient(gs.GetPeer(ctx, ttnpb.PeerInfo_ENTITY_REGISTRY, ids).Conn())
}

func (gs *GatewayServer) getContextForGatewayEUI(ctx context.Context, eui types.EUI64, opts ...grpc.CallOption) (context.Context, error) {
	if tenantID := tenant.FromContext(ctx); tenantID.TenantID != "" {
		return ctx, nil
	}
	ids, err := gs.getTenantRegistry(ctx, nil).GetIdentifiersForGatewayEUI(ctx, &ttipb.GetTenantIdentifiersForGatewayEUIRequest{
		EUI: eui,
	}, opts...)
	if err != nil {
		return nil, err
	}
	return tenant.NewContext(ctx, *ids), nil
}
