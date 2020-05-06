// Copyright © 2019 The Things Industries B.V.

package component

import (
	"context"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func (c *Component) initTenancy() {
	var fetcher tenant.Fetcher = tenant.FetcherFunc(func(ctx context.Context, ids *ttipb.TenantIdentifiers, fieldPaths ...string) (*ttipb.Tenant, error) {
		cc, err := c.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
		if err != nil {
			return nil, err
		}
		cli := ttipb.NewTenantRegistryClient(cc)
		creds := c.WithClusterAuth()
		return cli.Get(ctx, &ttipb.GetTenantRequest{
			TenantIdentifiers: *ids,
			FieldMask:         pbtypes.FieldMask{Paths: fieldPaths},
		}, creds)
	})
	if c.config.Tenancy.CacheTTL > 0 {
		fetcher = tenant.NewCachedFetcher(fetcher, c.config.Tenancy.CacheTTL, c.config.Tenancy.CacheTTL)
	} else {
		c.Logger().Warn("No tenant cache TTL configured")
	}
	fetcher = tenant.NewSingleFlightFetcher(fetcher)
	c.AddContextFiller(func(ctx context.Context) context.Context {
		return tenant.NewContextWithFetcher(ctx, fetcher)
	})
}
