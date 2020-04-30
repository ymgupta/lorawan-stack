// Copyright © 2019 The Things Industries B.V.

package tbsmetrics_test

import (
	"context"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

type mockRegisterer struct {
	ctx context.Context
	tbs ttipb.TbsServer
}

func (r *mockRegisterer) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_TENANT_BILLING_SERVER}
}

func (r *mockRegisterer) RegisterServices(s *grpc.Server) {
	ttipb.RegisterTbsServer(s, r.tbs)
}

func (r *mockRegisterer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttipb.RegisterTbsHandler(r.ctx, s, conn)
}

type mockTBS struct {
	ctx struct {
		Report context.Context
	}
	req struct {
		Report *ttipb.MeteringData
	}
	res struct {
		Report *types.Empty
	}
	err struct {
		Report error
	}
}

func (m *mockTBS) Report(ctx context.Context, data *ttipb.MeteringData) (*types.Empty, error) {
	m.ctx.Report, m.req.Report = ctx, data
	if m.res.Report != nil || m.err.Report != nil {
		return m.res.Report, m.err.Report
	}
	return ttnpb.Empty, nil
}

type mockConnProvider struct {
	c *component.Component
}

func (m *mockConnProvider) GetPeerConn(ctx context.Context, role ttnpb.ClusterRole) (*grpc.ClientConn, error) {
	return m.c.GetPeerConn(ctx, role, nil)
}

func mustHavePeer(ctx context.Context, c *component.Component, role ttnpb.ClusterRole) {
	for i := 0; i < 20; i++ {
		if _, err := c.GetPeer(ctx, role, nil); err == nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	panic("could not connect to peer")
}
