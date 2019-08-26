// Copyright © 2019 The Things Industries B.V.

package deviceclaimingserver_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/component"
	. "go.thethings.network/lorawan-stack/pkg/deviceclaimingserver"
	"go.thethings.network/lorawan-stack/pkg/deviceclaimingserver/redis"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestAuthorizeApplication(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	redisClient, redisFlush := test.NewRedis(t, "deviceclaimingserver_test", "applications", "authorized")
	defer redisFlush()
	defer redisClient.Close()
	authorizedApplicationsRegistry := &redis.AuthorizedApplicationRegistry{Redis: redisClient}

	c := component.MustNew(test.GetLogger(t), &component.Config{})
	test.Must(New(c, &Config{
		AuthorizedApplications: authorizedApplicationsRegistry,
	}))
	test.Must(c.Start(), nil)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.PeerInfo_DEVICE_CLAIMING_SERVER)

	client := ttnpb.NewEndDeviceClaimingServerClient(c.LoopbackConn())

	ids := ttnpb.ApplicationIdentifiers{
		ApplicationID: "foo",
	}

	// Application not authorized yet.
	_, err := client.UnauthorizeApplication(ctx, &ids)
	a.So(errors.IsNotFound(err), should.BeTrue)

	// Authorize application.
	_, err = client.AuthorizeApplication(ctx, &ttnpb.AuthorizeApplicationRequest{
		ApplicationIdentifiers: ids,
		APIKey:                 "test",
	})
	a.So(err, should.BeNil)

	// Authorize application again with new API key.
	_, err = client.AuthorizeApplication(ctx, &ttnpb.AuthorizeApplicationRequest{
		ApplicationIdentifiers: ids,
		APIKey:                 "test-new",
	})
	a.So(err, should.BeNil)

	// Unauthorize application.
	_, err = client.UnauthorizeApplication(ctx, &ids)
	a.So(err, should.BeNil)

	// Application not authorized anymore.
	_, err = client.UnauthorizeApplication(ctx, &ids)
	a.So(errors.IsNotFound(err), should.BeTrue)
}
