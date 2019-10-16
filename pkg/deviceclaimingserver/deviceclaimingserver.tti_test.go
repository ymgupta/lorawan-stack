// Copyright © 2019 The Things Industries B.V.

package deviceclaimingserver_test

import (
	"testing"

	"go.thethings.network/lorawan-stack/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/pkg/component/test"
	. "go.thethings.network/lorawan-stack/pkg/deviceclaimingserver"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/web/oauthclient"
)

func TestDeviceDeviceClaimingServer(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	conf := &component.Config{}
	c := componenttest.NewComponent(t, conf)

	test.Must(New(c, &Config{
		OAuth: oauthclient.Config{
			AuthorizeURL: "http://localhost/oauth/authorize",
			TokenURL:     "http://localhost/token",
			ClientID:     "test",
			ClientSecret: "test",
		},
	}))

	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_DEVICE_CLAIMING_SERVER)
}
