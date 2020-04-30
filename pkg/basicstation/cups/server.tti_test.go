// Copyright © 2019 The Things Industries B.V.

package cups

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
)

type mockTenantClientData struct {
	ctx struct {
		GetIdentifiersForGatewayEUI context.Context
	}
	req struct {
		GetIdentifiersForGatewayEUI *ttipb.GetTenantIdentifiersForGatewayEUIRequest
	}
	opts struct {
		GetIdentifiersForGatewayEUI []grpc.CallOption
	}
	res struct {
		GetIdentifiersForGatewayEUI *ttipb.TenantIdentifiers
	}
	err struct {
		GetIdentifiersForGatewayEUI error
	}
}

type mockTenantClient struct {
	mockTenantClientData
	ttipb.TenantRegistryClient
}

func (m *mockTenantClient) GetIdentifiersForGatewayEUI(ctx context.Context, in *ttipb.GetTenantIdentifiersForGatewayEUIRequest, opts ...grpc.CallOption) (*ttipb.TenantIdentifiers, error) {
	m.ctx.GetIdentifiersForGatewayEUI, m.req.GetIdentifiersForGatewayEUI, m.opts.GetIdentifiersForGatewayEUI = ctx, in, opts
	if m.err.GetIdentifiersForGatewayEUI != nil || m.res.GetIdentifiersForGatewayEUI != nil {
		return m.res.GetIdentifiersForGatewayEUI, m.err.GetIdentifiersForGatewayEUI
	}
	id := tenant.FromContext(test.Context())
	return &id, nil
}
