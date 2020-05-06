// Copyright © 2020 The Things Industries B.V.

package networkserver

import (
	"context"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

func init() {
	DefaultOptions = append(DefaultOptions,
		WithTenantOverrides(),
	)
}

func tenantConfigFromContext(ctx context.Context) (*ttipb.Configuration_Cluster_NetworkServer, bool) {
	conf, err := tenant.ConfigFromContext(ctx)
	if err != nil {
		return nil, false
	}
	nsConf := conf.GetDefaultCluster().GetNS()
	return nsConf, nsConf != nil
}

func WithTenantOverrides() Option {
	return func(ns *NetworkServer) {
		origNewDevAddr := ns.newDevAddr
		ns.newDevAddr = func(ctx context.Context, dev *ttnpb.EndDevice) types.DevAddr {
			conf, ok := tenantConfigFromContext(ctx)
			if !ok || len(conf.DevAddrPrefixes) == 0 {
				return origNewDevAddr(ctx, dev)
			}
			return makeNewDevAddrFunc(conf.DevAddrPrefixes...)(ctx, dev)
		}

		origDeduplicationWindow := ns.deduplicationWindow
		ns.deduplicationWindow = func(ctx context.Context) time.Duration {
			conf, ok := tenantConfigFromContext(ctx)
			if !ok || conf.DeduplicationWindow == nil {
				return origDeduplicationWindow(ctx)
			}
			return makeWindowDurationFunc(*conf.DeduplicationWindow)(ctx)
		}

		origCollectionWindow := ns.collectionWindow
		ns.collectionWindow = func(ctx context.Context) time.Duration {
			conf, ok := tenantConfigFromContext(ctx)
			if !ok || conf.DeduplicationWindow == nil || conf.CooldownWindow == nil {
				return origCollectionWindow(ctx)
			}
			return makeWindowDurationFunc(*conf.DeduplicationWindow + *conf.CooldownWindow)(ctx)
		}
	}
}
