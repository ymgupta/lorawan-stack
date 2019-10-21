// Copyright © 2019 The Things Industries B.V.

package tenantbillingserver

import (
	"context"

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Report implements ttipb.TbsServer.
func (tbs *TenantBillingServer) Report(ctx context.Context, data *ttipb.MeteringData) (*types.Empty, error) {
	if !billingRightsFromContext(ctx).report {
		return nil, errNoBillingRights
	}
	tenantFetcher, ok := tenant.FetcherFromContext(ctx)
	if !ok {
		panic("tenant fetcher not available")
	}
	for _, tenantData := range data.Tenants {
		tenant, err := tenantFetcher.FetchTenant(ctx, &tenantData.TenantIdentifiers, "attributes")
		if err != nil {
			return nil, err
		}
		for _, backend := range tbs.backends {
			err := backend.Report(ctx, tenant, tenantData.Totals)
			if err != nil {
				return nil, err
			}
		}
	}
	return ttnpb.Empty, nil
}
