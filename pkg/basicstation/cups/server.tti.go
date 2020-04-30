// Copyright © 2019 The Things Industries B.V.

package cups

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
	return func(s *Server) {
		s.tenantRegistry = registry
	}
}

func (s *Server) getTenantRegistry(ctx context.Context) (ttipb.TenantRegistryClient, error) {
	if s.tenantRegistry != nil {
		return s.tenantRegistry, nil
	}
	cc, err := s.component.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		return nil, err
	}
	return ttipb.NewTenantRegistryClient(cc), nil
}

func (s *Server) getContextForGatewayEUI(ctx context.Context, eui types.EUI64, opts ...grpc.CallOption) (context.Context, error) {
	if tenantID := tenant.FromContext(ctx); tenantID.TenantID != "" {
		return ctx, nil
	}
	registry, err := s.getTenantRegistry(ctx)
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
