// Copyright © 2020 The Things Industries B.V.

package cluster

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v7"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// RedisClaimRegistry is an implementation of the ClaimRegistry that uses Redis
// for peristence.
type RedisClaimRegistry struct {
	Redis  *ttnredis.Client
	PeerID string

	mu             sync.Mutex
	startKeepAlive sync.Once
	activeKeys     map[string]struct{}
}

// KeepAlive will continuously extend the TTLs of Redis keys used by the claim
// registry. This ensures that no keys are leaked if the program finishes or
// crashes.
func (r *RedisClaimRegistry) KeepAlive(ctx context.Context, ttl time.Duration) {
	ticker := time.NewTicker(ttl / 2)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			// Final cleanup of Redis registry.
			r.mu.Lock()
			defer r.mu.Unlock()
			_, err := r.Redis.Pipelined(func(tx redis.Pipeliner) error {
				for k := range r.activeKeys {
					tx.Del(k)
				}
				return nil
			})
			if err != nil {
				log.FromContext(ctx).WithError(err).Error("Failed to clear cluster claims")
			} else {
				for k := range r.activeKeys {
					delete(r.activeKeys, k)
				}
			}
			return
		case <-ticker.C:
			// Periodic TTL extension.
			_, err := r.Redis.Pipelined(func(tx redis.Pipeliner) error {
				r.mu.Lock()
				defer r.mu.Unlock()
				for k := range r.activeKeys {
					tx.Expire(k, ttl)
				}
				return nil
			})
			if err != nil {
				log.FromContext(ctx).WithError(err).Error("Failed to extend cluster claim expiry")
			}
		}
	}
}

func (r *RedisClaimRegistry) setKeysActive(keys ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.activeKeys == nil {
		r.activeKeys = make(map[string]struct{})
	}
	for _, key := range keys {
		r.activeKeys[key] = struct{}{}
	}
}

func (r *RedisClaimRegistry) sMembers(ctx context.Context, ids ttnpb.Identifiers) map[string]string {
	out := map[string]string{
		r.Redis.Key(r.PeerID, fmt.Sprintf("%s-ids", ids.EntityType())): unique.ID(ctx, ids),
	}
	if ids, ok := ids.Identifiers().(*ttnpb.GatewayIdentifiers); ok && ids.EUI != nil {
		out[r.Redis.Key(r.PeerID, "gateway-euis")] = ids.EUI.String()
	}
	return out
}

// Claim claims the given identifiers on the current peer.
func (r *RedisClaimRegistry) Claim(ctx context.Context, ids ttnpb.Identifiers) error {
	members := r.sMembers(ctx, ids)
	_, err := r.Redis.TxPipelined(func(tx redis.Pipeliner) error {
		for k, m := range members {
			tx.SAdd(k, m)
			r.setKeysActive(k)
		}
		return nil
	})
	return err
}

// Unclaim releases the claim the current peer has on the given identifiers.
func (r *RedisClaimRegistry) Unclaim(ctx context.Context, ids ttnpb.Identifiers) error {
	members := r.sMembers(ctx, ids)
	_, err := r.Redis.TxPipelined(func(tx redis.Pipeliner) error {
		for k, m := range members {
			tx.SRem(k, m)
			r.setKeysActive(k)
		}
		return nil
	})
	return err
}

// GetPeerID looks up which from the given candidates has a claim on the given identifiers.
func (r *RedisClaimRegistry) GetPeerID(ctx context.Context, ids ttnpb.Identifiers, candidateIDs ...string) (string, error) {
	results := make([]*redis.BoolCmd, len(candidateIDs))
	_, err := r.Redis.ReadOnlyClient().Pipelined(func(tx redis.Pipeliner) error {
		for i, candidateID := range candidateIDs {
			if ids, ok := ids.Identifiers().(*ttnpb.GatewayIdentifiers); ok && ids.EUI != nil {
				results[i] = tx.SIsMember(
					r.Redis.Key(candidateID, "gateway-euis"),
					ids.EUI.String(),
				)
			} else {
				results[i] = tx.SIsMember(
					r.Redis.Key(candidateID, fmt.Sprintf("%s-ids", ids.EntityType())),
					unique.ID(ctx, ids),
				)
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	for i, result := range results {
		if result.Val() {
			return candidateIDs[i], nil
		}
	}
	return "", errPeerUnavailable
}
