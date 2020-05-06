// Copyright © 2020 The Things Industries B.V.

package cluster

import (
	"context"
	"sync"
	"time"

	"github.com/bluele/gcache"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"golang.org/x/sync/singleflight"
)

// ClaimRegistryCache is a cache on top of a ClaimRegistry.
type ClaimRegistryCache struct {
	ClaimRegistry
	Size int
	TTL  time.Duration

	initOnce sync.Once
	cache    gcache.Cache

	singleflight singleflight.Group
}

type getPeerResult struct {
	peerID string
	err    error
}

func (c *ClaimRegistryCache) init() {
	c.initOnce.Do(func() {
		size := c.Size
		if size == 0 {
			size = 16 * 1024
		}
		ttl := c.TTL
		if ttl == 0 {
			ttl = time.Minute
		}
		c.cache = gcache.New(size).LFU().Expiration(ttl).Build()
	})
}

// Invalidate invalidates the entire cache.
func (c *ClaimRegistryCache) Invalidate() {
	c.init()
	c.cache.Purge()
}

type invalidator interface {
	Invalidate()
}

// GetPeerID tries to get the peer ID from cache. If it's not cached, it will
// call the underlying ClaimRegistry and cache the result.
func (c *ClaimRegistryCache) GetPeerID(ctx context.Context, ids ttnpb.Identifiers, candidateIDs ...string) (string, error) {
	c.init()
	uid := unique.ID(ctx, ids)
	cachedID, err := c.cache.Get(uid)
	if err == nil {
		for _, candidateID := range candidateIDs {
			if cachedID == candidateID {
				return candidateID, nil
			}
		}
	}
	peerID, err, _ := c.singleflight.Do(uid, func() (interface{}, error) {
		peerID, err := c.ClaimRegistry.GetPeerID(ctx, ids, candidateIDs...)
		if err != nil {
			return "", err
		}
		c.cache.Set(uid, peerID)
		return peerID, nil
	})
	return peerID.(string), err
}
