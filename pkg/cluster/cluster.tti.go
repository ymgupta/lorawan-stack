// Copyright © 2020 The Things Industries B.V.

package cluster

import (
	"context"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

var errInvalidDiscoveryMode = errors.DefineInvalidArgument("discovery_mode", "invalid discovery mode")

func init() {
	CustomNew = func(ctx context.Context, config *config.Cluster, options ...Option) (Cluster, error) {
		switch strings.ToLower(config.DiscoveryMode) {
		case "dns":
			return newDNS(ctx, config, options...)
		case "":
			return defaultNew(ctx, config, options...)
		default:
			return nil, errInvalidDiscoveryMode
		}
	}
}

type peersByName []*peer

func (a peersByName) Len() int           { return len(a) }
func (a peersByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a peersByName) Less(i, j int) bool { return a[i].Name() < a[j].Name() }
