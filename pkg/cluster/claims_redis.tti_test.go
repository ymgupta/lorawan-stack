// Copyright © 2020 The Things Industries B.V.

package cluster_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestRedisClaimRegistry(t *testing.T) {
	ctx := test.Context()
	a := assertions.New(t)

	client, flush := test.NewRedis(t)
	defer flush()
	defer client.Close()

	reg1 := &cluster.RedisClaimRegistry{Redis: client, PeerID: "peer1"}
	err := reg1.Claim(ctx, &ttnpb.GatewayIdentifiers{GatewayID: "foo"})
	a.So(err, should.BeNil)

	reg2 := &cluster.RedisClaimRegistry{Redis: client, PeerID: "peer2"}
	err = reg2.Claim(ctx, &ttnpb.GatewayIdentifiers{GatewayID: "bar"})
	a.So(err, should.BeNil)

	peerID, err := reg1.GetPeerID(ctx, &ttnpb.GatewayIdentifiers{GatewayID: "bar"}, "peer1", "peer2")
	a.So(err, should.BeNil)
	a.So(peerID, should.Equal, "peer2")

	err = reg2.Unclaim(ctx, &ttnpb.GatewayIdentifiers{GatewayID: "bar"})
	a.So(err, should.BeNil)

	peerID, err = reg1.GetPeerID(ctx, &ttnpb.GatewayIdentifiers{GatewayID: "bar"}, "peer1", "peer2")
	a.So(err, should.NotBeNil)
}
