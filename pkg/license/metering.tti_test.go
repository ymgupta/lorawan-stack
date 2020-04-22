// Copyright © 2019 The Things Industries B.V.

package license_test

import (
	"context"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	. "go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
)

type mockIS struct {
	ttipb.UnimplementedTenantRegistryServer
	ctx struct {
		List              context.Context
		GetRegistryTotals context.Context
	}
	req struct {
		List              *ttipb.ListTenantsRequest
		GetRegistryTotals *ttipb.GetTenantRegistryTotalsRequest
	}
	res struct {
		List              *ttipb.Tenants
		GetRegistryTotals *ttipb.TenantRegistryTotals
	}
	err struct {
		List              error
		GetRegistryTotals error
	}
}

func (m *mockIS) List(ctx context.Context, req *ttipb.ListTenantsRequest) (*ttipb.Tenants, error) {
	m.ctx.List, m.req.List = ctx, req
	return m.res.List, m.err.List
}

func (m *mockIS) GetRegistryTotals(ctx context.Context, req *ttipb.GetTenantRegistryTotalsRequest) (*ttipb.TenantRegistryTotals, error) {
	m.ctx.GetRegistryTotals, m.req.GetRegistryTotals = ctx, req
	return m.res.GetRegistryTotals, m.err.GetRegistryTotals
}

type mockCluster struct {
	test.MockCluster
}

func (c *mockCluster) GetPeerConn(ctx context.Context, role ttnpb.ClusterRole) (*grpc.ClientConn, error) {
	return c.MockCluster.GetPeerConn(ctx, role, nil)
}

func newISPeer(ctx context.Context, is interface {
	ttipb.TenantRegistryServer
}) cluster.Peer {
	return test.Must(test.NewGRPCServerPeer(ctx, is, ttipb.RegisterTenantRegistryServer)).(cluster.Peer)
}

type mockReporter struct {
	ctx  context.Context
	data *ttipb.MeteringData
	err  error
}

func (r *mockReporter) Report(ctx context.Context, data *ttipb.MeteringData) error {
	r.ctx, r.data = ctx, data
	return r.err
}

func TestMetering(t *testing.T) {
	ctx := test.Context()
	a := assertions.New(t)

	mockIS := &mockIS{}
	mockCluster := &mockCluster{}
	mockCluster.AuthFunc = func() grpc.CallOption {
		return grpc.PerRPCCredentials(nil)
	}
	mockCluster.GetPeerFunc = func(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) (cluster.Peer, error) {
		return newISPeer(ctx, mockIS), nil
	}
	mockReporter := &mockReporter{}

	_, err := NewMeteringSetup(ctx, &ttipb.MeteringConfiguration{}, mockCluster)
	a.So(err, should.NotBeNil)

	_, err = NewMeteringSetup(ctx, &ttipb.MeteringConfiguration{
		Metering: &ttipb.MeteringConfiguration_AWS_{
			AWS: &ttipb.MeteringConfiguration_AWS{},
		},
	}, mockCluster)
	a.So(err, should.BeNil)

	tenMinutes := 10 * time.Minute
	s, err := NewMeteringSetup(ctx, &ttipb.MeteringConfiguration{
		Metering: &ttipb.MeteringConfiguration_Prometheus_{
			Prometheus: &ttipb.MeteringConfiguration_Prometheus{},
		},
		OnSuccess: &ttipb.LicenseUpdate{
			ExtendValidUntil: &tenMinutes,
		},
	}, mockCluster)
	a.So(err, should.BeNil)

	s.ReplaceReporter(mockReporter)

	license := ttipb.License{
		ValidUntil: time.Now().Add(time.Minute),
	}
	a.So(s.Apply(license), should.Resemble, license)

	t.Run("Empty Tenants List", func(t *testing.T) {
		a := assertions.New(t)

		mockIS.res.List = &ttipb.Tenants{}
		mockIS.res.GetRegistryTotals = &ttipb.TenantRegistryTotals{
			Applications:  1,
			Clients:       2,
			EndDevices:    3,
			Gateways:      4,
			Organizations: 5,
			Users:         6,
		}

		err = s.CollectAndReport(ctx)
		a.So(err, should.BeNil)

		if a.So(mockReporter.data.Tenants, should.HaveLength, 1) {
			a.So(mockReporter.data.Tenants[0].GetTenantID(), should.Equal, "")
			a.So(mockReporter.data.Tenants[0].GetTotals(), should.Resemble, mockIS.res.GetRegistryTotals)
		}
	})

	t.Run("Single Tenant", func(t *testing.T) {
		a := assertions.New(t)

		mockIS.res.List = &ttipb.Tenants{Tenants: []*ttipb.Tenant{
			{TenantIdentifiers: ttipb.TenantIdentifiers{TenantID: "foo"}},
			{TenantIdentifiers: ttipb.TenantIdentifiers{TenantID: "bar"}},
			{TenantIdentifiers: ttipb.TenantIdentifiers{TenantID: "baz"}},
		}}
		mockIS.res.GetRegistryTotals = &ttipb.TenantRegistryTotals{
			Applications:  1,
			Clients:       2,
			EndDevices:    3,
			Gateways:      4,
			Organizations: 5,
			Users:         6,
		}

		err = s.CollectAndReport(ctx)
		a.So(err, should.BeNil)

		a.So(mockReporter.data.Tenants, should.HaveLength, 3)
	})

	a.So(s.Apply(license).ValidUntil, should.NotEqual, license.ValidUntil)
}
