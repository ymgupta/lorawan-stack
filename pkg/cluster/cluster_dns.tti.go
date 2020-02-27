// Copyright © 2020 The Things Industries B.V.

package cluster

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/golang/groupcache/consistenthash"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcclient"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func newDNS(ctx context.Context, config *config.Cluster, options ...Option) (Cluster, error) {
	c := &dnsCluster{
		cluster: &cluster{
			ctx:   ctx,
			tls:   config.TLS,
			peers: make(map[string]*peer),
		},
		resolver:      net.DefaultResolver,
		peerDiscovery: make(map[string][]ttnpb.ClusterRole),
	}

	if err := c.loadKeys(ctx, config.Keys...); err != nil {
		return nil, err
	}

	c.self = &peer{
		name:   config.Name,
		target: config.Address,
	}
	if c.self.name == "" {
		c.self.name, _ = os.Hostname()
	}

	for _, option := range options {
		option.apply(c.cluster)
	}

	c.addPeerDiscovery(config.IdentityServer, ttnpb.ClusterRole_ACCESS, ttnpb.ClusterRole_ENTITY_REGISTRY)
	c.addPeerDiscovery(config.GatewayServer, ttnpb.ClusterRole_GATEWAY_SERVER)
	c.addPeerDiscovery(config.NetworkServer, ttnpb.ClusterRole_NETWORK_SERVER)
	c.addPeerDiscovery(config.ApplicationServer, ttnpb.ClusterRole_APPLICATION_SERVER)
	c.addPeerDiscovery(config.JoinServer, ttnpb.ClusterRole_JOIN_SERVER)
	c.addPeerDiscovery(config.CryptoServer, ttnpb.ClusterRole_CRYPTO_SERVER)
	c.addPeerDiscovery(config.PacketBrokerAgent, ttnpb.ClusterRole_PACKET_BROKER_AGENT)

	for _, join := range config.Join {
		c.addPeerDiscovery(join, ttnpb.ClusterRole_NONE)
	}

	return c, nil
}

type dnsResolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
	LookupSRV(ctx context.Context, service, proto, name string) (cname string, addrs []*net.SRV, err error)
}

type dnsCluster struct {
	*cluster

	resolver dnsResolver

	nextDiscovery   *time.Ticker
	cancelDiscovery chan struct{}
	peerDiscovery   map[string][]ttnpb.ClusterRole

	peerMu           sync.RWMutex
	byRole           map[ttnpb.ClusterRole][]*peer
	consistentHashes map[ttnpb.ClusterRole]*consistenthash.Map
}

func (c *dnsCluster) addPeerDiscovery(target string, roles ...ttnpb.ClusterRole) {
	if target == "" {
		return
	}
	c.peerDiscovery[target] = roles
}

func (c *dnsCluster) updatePeers(ctx context.Context) {
	logger := log.FromContext(ctx)
	peers := make(map[string]*peer)

	// Discover peers with DNS lookups.
	for address, roles := range c.peerDiscovery {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			host, port = address, ""
		}
		var addresses []string
		if port != "" { // Port is already known.
			addresses = []string{address}
		} else { // Port is unknown, discover hostnames and ports.
			_, records, err := c.resolver.LookupSRV(ctx, "", "", host)
			if err != nil {
				logger.WithField("address", address).WithError(err).Error("DNS lookup failed")
				continue
			}
			for _, record := range records {
				addresses = append(addresses, fmt.Sprintf("%s:%d", record.Target, record.Port))
			}
		}
		for _, address := range addresses {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				logger.WithField("address", address).WithError(err).Error("DNS lookup failed")
				continue
			}
			records, err := c.resolver.LookupIPAddr(ctx, host)
			if err != nil {
				logger.WithField("address", address).WithError(err).Error("DNS lookup failed")
				continue
			}
			for _, record := range records {
				address := net.JoinHostPort(record.String(), port)
				peers[address] = &peer{
					name:   address,
					roles:  roles,
					target: address,
				}
			}
		}
	}

	options := rpcclient.DefaultDialOptions(c.ctx)
	if c.tls {
		options = append(options, grpc.WithTransportCredentials(credentials.NewTLS(c.tlsConfig)))
	} else {
		options = append(options, grpc.WithInsecure())
	}

	// Re-use existing peers, connect to new peers.
	for name, peer := range peers {
		if existing, ok := c.peers[name]; ok && existing.target == peer.target {
			peers[name] = existing
			continue
		}
		peer.ctx, peer.cancel = context.WithCancel(ctx)
		logger := logger.WithFields(log.Fields(
			"target", peer.target,
			"name", peer.Name(),
			"roles", peer.Roles(),
		))
		logger.Debug("Connecting to peer...")
		peer.conn, peer.connErr = grpc.DialContext(peer.ctx, peer.target, options...)
		if peer.connErr != nil {
			logger.WithError(peer.connErr).Error("Failed to connect to peer")
		} else {
			logger.Debug("Connected to peer")
		}
	}

	// Construct lookup maps.
	byRole := make(map[ttnpb.ClusterRole][]*peer)
	consistentHashes := make(map[ttnpb.ClusterRole]*consistenthash.Map)
	for name, peer := range peers {
		for _, role := range peer.roles {
			byRole[role] = append(byRole[role], peer)
			hashMap, ok := consistentHashes[role]
			if !ok {
				hashMap = consistenthash.New(8, nil)
				consistentHashes[role] = hashMap
			}
			hashMap.Add(name)
		}
	}
	for _, peers := range byRole {
		sort.Sort(peersByName(peers))
	}

	// Collect old peers and disconnect asynchronously.
	var oldPeers []*peer
	for name, peer := range c.peers {
		if _, stillExists := peers[name]; !stillExists {
			logger.WithFields(log.Fields(
				"target", peer.target,
				"name", peer.Name(),
				"roles", peer.Roles(),
			)).Debug("Schedule peer for disconnect")
			oldPeers = append(oldPeers, peer)
		}
	}
	if len(oldPeers) > 0 {
		// Give pending RPCs 10 seconds to finish, then close the conns.
		time.AfterFunc(10*time.Second, func() {
			for _, peer := range oldPeers {
				if peer.conn != nil {
					peer.conn.Close()
				}
				if peer.cancel != nil {
					peer.cancel()
				}
			}
		})
	}

	// Replace the current state of the cluster.
	c.peerMu.Lock()
	c.peers = peers
	c.byRole = byRole
	c.consistentHashes = consistentHashes
	c.peerMu.Unlock()
}

func (c *dnsCluster) Join() (err error) {
	c.updatePeers(c.ctx)
	c.nextDiscovery = time.NewTicker(10 * time.Second)
	c.cancelDiscovery = make(chan struct{})
	go func() {
		for {
			select {
			case <-c.cancelDiscovery:
				return
			case <-c.nextDiscovery.C:
				c.updatePeers(c.ctx)
			}
		}
	}()
	return nil
}

func (c *dnsCluster) Leave() error {
	c.nextDiscovery.Stop()
	close(c.cancelDiscovery)
	for _, peer := range c.peers {
		if peer.conn != nil {
			peer.conn.Close()
		}
		if peer.cancel != nil {
			peer.cancel()
		}
	}
	return nil
}

func (c *dnsCluster) GetPeers(ctx context.Context, role ttnpb.ClusterRole) ([]Peer, error) {
	c.peerMu.RLock()
	defer c.peerMu.RUnlock()

	peers, ok := c.byRole[role]
	if !ok || len(peers) == 0 {
		return nil, nil
	}
	peerInterfaces := make([]Peer, len(peers))
	for i, peer := range peers {
		peerInterfaces[i] = peer
	}
	return peerInterfaces, nil
}

func (c *dnsCluster) GetPeer(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) (Peer, error) {
	role = overridePeerRole(ctx, role, ids)
	roleString := strings.Title(strings.Replace(role.String(), "_", " ", -1))

	c.peerMu.RLock()
	defer c.peerMu.RUnlock()

	if ids == nil {
		matches := c.byRole[role]
		if len(matches) == 0 {
			return nil, errPeerUnavailable.WithAttributes("cluster_role", roleString)
		}
		return matches[rand.Intn(len(matches))], nil
	}

	switch role {
	case ttnpb.ClusterRole_GATEWAY_SERVER:
		matches := c.byRole[role]
		switch len(matches) {
		case 0:
			return nil, errPeerUnavailable.WithAttributes("cluster_role", roleString)
		case 1:
			return matches[0], nil
		default:
			// TODO: ID Claim lookup (https://github.com/TheThingsIndustries/lorawan-stack/issues/1970).
			return matches[0], nil
		}
	default:
		hashMap := c.consistentHashes[role]
		if hashMap == nil {
			return nil, errPeerUnavailable.WithAttributes("cluster_role", roleString)
		}
		key := ids.EntityType() + ":" + unique.ID(ctx, ids)
		peerID := hashMap.Get(key)
		if peerID == "" {
			return nil, errPeerUnavailable.WithAttributes("cluster_role", roleString)
		}
		log.FromContext(ctx).Debugf("Hashed %s to %s", key, peerID)
		peer := c.peers[peerID]
		if peer == nil {
			return nil, errPeerUnavailable.WithAttributes("cluster_role", roleString)
		}
		return peer, nil
	}
}

func (c *dnsCluster) GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) (*grpc.ClientConn, error) {
	peer, err := c.GetPeer(ctx, role, ids)
	if err != nil {
		return nil, err
	}
	return peer.Conn()
}

func (c *dnsCluster) ClaimIDs(ctx context.Context, ids ttnpb.Identifiers) error {
	// TODO: ID Claim lookup (https://github.com/TheThingsIndustries/lorawan-stack/issues/1970).
	return nil
}

func (c *dnsCluster) UnclaimIDs(ctx context.Context, ids ttnpb.Identifiers) error {
	// TODO: ID Claim lookup (https://github.com/TheThingsIndustries/lorawan-stack/issues/1970).
	return nil
}
