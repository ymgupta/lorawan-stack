// Copyright © 2020 The Things Industries B.V.

package cluster

import (
	"context"
	"strings"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// ClaimRegistryConfig represents configuration for the cluster claim registry.
type ClaimRegistryConfig struct {
	Backend string `name:"backend" description:"ID claiming backend (redis)"`
	Cache   struct {
		Size int           `name:"size" description:"Maximum size of local claims cache"`
		TTL  time.Duration `name:"ttl" description:"TTL of locally cached claims"`
	} `name:"cache"`
	PeerID string        `name:"-"`
	Redis  *redis.Client `name:"-"`
}

// ClaimRegistry is the interface that is used for working with ID claims in a cluster.
type ClaimRegistry interface {
	Claim(ctx context.Context, ids ttnpb.Identifiers) error
	Unclaim(ctx context.Context, ids ttnpb.Identifiers) error
	GetPeerID(ctx context.Context, ids ttnpb.Identifiers, candidateIDs ...string) (string, error)
}

var errInvalidClaimingBackend = errors.DefineInvalidArgument("claiming_backend", "invalid ID claiming backend")

// NewClaimRegistry instantiates a new cluster claim registry based on the given config.
func NewClaimRegistry(ctx context.Context, config *ClaimRegistryConfig) (ClaimRegistry, error) {
	var reg ClaimRegistry
	switch strings.ToLower(config.Backend) {
	case "redis":
		redisRegistry := &RedisClaimRegistry{
			Redis:  config.Redis,
			PeerID: config.PeerID,
		}
		go redisRegistry.KeepAlive(ctx, time.Hour)
		reg = redisRegistry
	default:
		return nil, errInvalidClaimingBackend.New()
	}
	reg = &ClaimRegistryCache{
		ClaimRegistry: reg,
		Size:          config.Cache.Size,
		TTL:           config.Cache.TTL,
	}
	return reg, nil
}
