// Copyright © 2019 The Things Industries B.V.

package tenantbillingserver_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/tenantbillingserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestAuth(t *testing.T) {
	withTBS(t, tenantbillingserver.Config{
		ReporterAddressRegexps: []string{"pipe"},
	},
		func(t *testing.T, c *component.Component) {
			a := assertions.New(t)
			ctx := test.Context()

			cc, err := c.GetPeerConn(ctx, ttnpb.ClusterRole_TENANT_BILLING_SERVER, nil)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}

			client := ttipb.NewTbsClient(cc)

			res, err := client.Report(ctx, &ttipb.MeteringData{})
			a.So(err, should.BeNil)
			a.So(res, should.NotBeNil)
		},
	)

	withTBS(t, tenantbillingserver.Config{
		ReporterAddressRegexps: []string{"never-valid-address"},
	},
		func(t *testing.T, c *component.Component) {
			a := assertions.New(t)
			ctx := test.Context()

			cc, err := c.GetPeerConn(ctx, ttnpb.ClusterRole_TENANT_BILLING_SERVER, nil)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}

			client := ttipb.NewTbsClient(cc)

			_, err = client.Report(ctx, &ttipb.MeteringData{})
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
		},
	)
}

func withTBS(t *testing.T, cfg tenantbillingserver.Config, testFunc func(*testing.T, *component.Component)) {
	a := assertions.New(t)

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
		},
	})

	s, err := tenantbillingserver.New(c, &cfg)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(s, should.NotBeNil)

	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(c.Context(), c, ttnpb.ClusterRole_TENANT_BILLING_SERVER)

	testFunc(t, c)
}
