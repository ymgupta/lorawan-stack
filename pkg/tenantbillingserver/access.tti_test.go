// Copyright © 2019 The Things Industries B.V.

package tenantbillingserver_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/pkg/component/test"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/tenantbillingserver"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestAuth(t *testing.T) {
	withTBS(t, tenantbillingserver.Config{
		ReporterAddressRegexps: []string{"pipe"},
	},
		func(t *testing.T, a *assertions.Assertion, c *component.Component) {
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
		func(t *testing.T, a *assertions.Assertion, c *component.Component) {
			ctx := test.Context()

			cc, err := c.GetPeerConn(ctx, ttnpb.ClusterRole_TENANT_BILLING_SERVER, nil)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}

			client := ttipb.NewTbsClient(cc)

			res, err := client.Report(ctx, &ttipb.MeteringData{})
			a.So(err, should.NotBeNil)
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
			a.So(res, should.BeNil)
		},
	)
}

func withTBS(t *testing.T, cfg tenantbillingserver.Config, testFunc func(*testing.T, *assertions.Assertion, *component.Component)) {
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
	a.So(s, should.NotBeNil)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(c.Context(), c, ttnpb.ClusterRole_TENANT_BILLING_SERVER)

	testFunc(t, a, c)
}
