// Copyright © 2019 The Things Industries B.V.

package tbsmetrics_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/pkg/component/test"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/license/tbsmetrics"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestReporter(t *testing.T) {
	ctx := test.Context()
	a := assertions.New(t)

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
		},
	})

	reporter, err := tbsmetrics.New(&ttipb.MeteringConfiguration_TenantBillingServer{}, &mockConnProvider{c})
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	tbs := &mockTBS{}
	tbs.res.Report = ttnpb.Empty
	tbs.err.Report = nil

	c.RegisterGRPC(&mockRegisterer{ctx, tbs})
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_TENANT_BILLING_SERVER)

	data := &ttipb.MeteringData{
		Tenants: []*ttipb.MeteringData_TenantMeteringData{
			{
				TenantIdentifiers: ttipb.TenantIdentifiers{
					TenantID: "test-tenant",
				},
				Totals: &ttipb.TenantRegistryTotals{
					Applications:  1,
					Clients:       2,
					EndDevices:    3,
					Gateways:      4,
					Organizations: 5,
					Users:         6,
				},
			},
		},
	}

	err = reporter.Report(ctx, data)
	a.So(err, should.BeNil)
	a.So(tbs.req.Report, should.Resemble, data)
}
