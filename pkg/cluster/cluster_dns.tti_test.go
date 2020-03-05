// Copyright © 2020 The Things Industries B.V.

package cluster_test

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

type mockResolver struct {
	mu      sync.RWMutex
	ipAddrs map[string][]net.IPAddr
	srvs    map[string][]*net.SRV
	errs    map[string]error
}

func newMockResolver() *mockResolver {
	return &mockResolver{
		ipAddrs: make(map[string][]net.IPAddr),
		srvs:    make(map[string][]*net.SRV),
		errs:    make(map[string]error),
	}
}

func (r *mockResolver) setIP(name string, ips ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(ips) == 0 {
		delete(r.ipAddrs, name)
	} else {
		addrs := make([]net.IPAddr, len(ips))
		for i, ip := range ips {
			addrs[i] = net.IPAddr{IP: net.ParseIP(ip)}
		}
		r.ipAddrs[name] = addrs
	}
}

func (r *mockResolver) setSRV(name string, hostports ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(hostports) == 0 {
		delete(r.srvs, name)
	} else {
		addrs := make([]*net.SRV, len(hostports))
		for i, hostport := range hostports {
			host, portStr, _ := net.SplitHostPort(hostport)
			port, _ := strconv.ParseUint(portStr, 10, 16)
			addrs[i] = &net.SRV{
				Target: host,
				Port:   uint16(port),
			}
		}
		r.srvs[name] = addrs
	}
}

func (r *mockResolver) setErr(name string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err == nil {
		delete(r.errs, name)
	} else {
		r.errs[name] = err
	}
}

func (r *mockResolver) LookupIPAddr(ctx context.Context, name string) ([]net.IPAddr, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ipAddrs[name], r.errs[name]
}

func (r *mockResolver) LookupSRV(ctx context.Context, _, _, name string) (cname string, addrs []*net.SRV, err error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return "", r.srvs[name], r.errs[name]
}

func TestDNSCluster(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()
	ctx = log.NewContext(ctx, test.GetLogger(t))

	c, err := cluster.New(ctx, &cluster.Config{
		DiscoveryMode:     "DNS",
		IdentityServer:    "is.cluster.local:1885",
		GatewayServer:     "gs.cluster.local",
		PacketBrokerAgent: "pba.cluster.local",
	})
	a.So(err, should.BeNil)

	r := newMockResolver()
	cluster.SetResolver(c, r)

	r.setErr("pba.cluster.local", fmt.Errorf("not found"))

	_, err = c.GetPeer(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	a.So(err, should.NotBeNil)

	t.Run("Single IS", func(t *testing.T) {
		a := assertions.New(t)
		ctx := log.NewContext(ctx, test.GetLogger(t))

		r.setIP("is.cluster.local", "10.0.0.1")

		cluster.UpdatePeers(ctx, c)

		peers, err := c.GetPeers(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY)
		a.So(err, should.BeNil)
		if a.So(peers, should.HaveLength, 1) {
			a.So(peers[0].Name(), should.Equal, "10.0.0.1:1885")
		}

		peer, err := c.GetPeer(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
		a.So(err, should.BeNil)
		a.So(peer, should.Equal, peers[0])

		peer, err = c.GetPeer(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, &ttnpb.ApplicationIdentifiers{ApplicationID: "foo"})
		a.So(err, should.BeNil)
		a.So(peer, should.Equal, peers[0])

		peers, err = c.GetPeers(ctx, ttnpb.ClusterRole_GATEWAY_SERVER)
		a.So(err, should.BeNil)
		a.So(peers, should.HaveLength, 0)
	})

	t.Run("Second IS", func(t *testing.T) {
		a := assertions.New(t)
		ctx := log.NewContext(ctx, test.GetLogger(t))

		r.setIP("is.cluster.local", "10.0.0.1", "10.0.0.2")

		cluster.UpdatePeers(ctx, c)

		peers, err := c.GetPeers(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY)
		a.So(err, should.BeNil)
		if a.So(peers, should.HaveLength, 2) {
			a.So(peers[0].Name(), should.Equal, "10.0.0.1:1885")
			a.So(peers[1].Name(), should.Equal, "10.0.0.2:1885")
		}

		peers, err = c.GetPeers(ctx, ttnpb.ClusterRole_GATEWAY_SERVER)
		a.So(err, should.BeNil)
		a.So(peers, should.HaveLength, 0)
	})

	t.Run("Add GSs", func(t *testing.T) {
		a := assertions.New(t)
		ctx := log.NewContext(ctx, test.GetLogger(t))

		r.setIP("gs1.cluster.local", "10.0.1.1")
		r.setIP("gs2.cluster.local", "10.0.1.2")
		r.setSRV("gs.cluster.local", "gs1.cluster.local:1885", "gs2.cluster.local:1885")

		cluster.UpdatePeers(ctx, c)

		peers, err := c.GetPeers(ctx, ttnpb.ClusterRole_GATEWAY_SERVER)
		a.So(err, should.BeNil)
		if a.So(peers, should.HaveLength, 2) {
			a.So(peers[0].Name(), should.Equal, "gs1")
			a.So(peers[1].Name(), should.Equal, "gs2")
		}
	})

	t.Run("Remove IS", func(t *testing.T) {
		a := assertions.New(t)
		ctx := log.NewContext(ctx, test.GetLogger(t))

		r.setIP("is.cluster.local", "10.0.0.1")

		cluster.UpdatePeers(ctx, c)

		peers, err := c.GetPeers(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY)
		a.So(err, should.BeNil)
		if a.So(peers, should.HaveLength, 1) {
			a.So(peers[0].Name(), should.Equal, "10.0.0.1:1885")
		}
	})
}
