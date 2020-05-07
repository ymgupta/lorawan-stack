// Copyright © 2019 The Things Industries B.V.

package gatewayserver

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/grpc"
)

// WithTenantRegistry overrides the CUPS server's tenant registry.
func WithTenantRegistry(registry ttipb.TenantRegistryClient) Option {
	return func(s *GatewayServer) {
		s.tenantRegistry = registry
	}
}

func (gs *GatewayServer) getTenantRegistry(ctx context.Context) (ttipb.TenantRegistryClient, error) {
	if gs.tenantRegistry != nil {
		return gs.tenantRegistry, nil
	}
	cc, err := gs.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		return nil, err
	}
	return ttipb.NewTenantRegistryClient(cc), nil
}

func (gs *GatewayServer) getContextForGatewayEUI(ctx context.Context, eui types.EUI64, opts ...grpc.CallOption) (context.Context, error) {
	if tenantID := tenant.FromContext(ctx); tenantID.TenantID != "" {
		return ctx, nil
	}
	registry, err := gs.getTenantRegistry(ctx)
	if err != nil {
		return nil, err
	}
	ids, err := registry.GetIdentifiersForGatewayEUI(ctx, &ttipb.GetTenantIdentifiersForGatewayEUIRequest{
		EUI: eui,
	}, opts...)
	if err != nil {
		return nil, err
	}
	return tenant.NewContext(ctx, *ids), nil
}
