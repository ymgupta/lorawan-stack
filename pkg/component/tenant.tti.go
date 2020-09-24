// Copyright © 2019 The Things Industries B.V.

package component

import (
	"context"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc/codes"
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
	if ttl := c.config.Tenancy.CacheTTL; ttl > 0 {
		fetcher = tenant.NewCachedFetcher(
			fetcher,
			tenant.StaticTTL(ttl),
			tenant.WithStaleDataForErrors(func(err error) bool {
				switch codes.Code(errors.Code(err)) {
				case codes.Canceled,
					codes.Unknown,
					codes.DeadlineExceeded,
					codes.ResourceExhausted,
					codes.Internal,
					codes.Unavailable,
					codes.DataLoss:
					return true
				default:
					return false
				}
			}),
		)
	} else {
		c.Logger().Warn("No tenant cache TTL configured")
	}
	fetcher = tenant.NewSingleFlightFetcher(fetcher)
	fetcher = tenant.NewCombinedFieldsFetcher(fetcher)
	c.AddContextFiller(func(ctx context.Context) context.Context {
		return tenant.NewContextWithFetcher(ctx, fetcher)
	})
}
