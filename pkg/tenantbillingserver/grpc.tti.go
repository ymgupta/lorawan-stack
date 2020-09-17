// Copyright © 2019 The Things Industries B.V.

package tenantbillingserver

import (
	"context"

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Report implements ttipb.TbsServer.
func (tbs *TenantBillingServer) Report(ctx context.Context, data *ttipb.MeteringData) (*types.Empty, error) {
	tenantFetcher, ok := tenant.FetcherFromContext(ctx)
	if !ok {
		panic("tenant fetcher not available")
	}
	logger := log.FromContext(ctx)
	for _, tenantData := range data.Tenants {
		tenant, err := tenantFetcher.FetchTenant(ctx, &tenantData.TenantIdentifiers, "billing", "state")
		if err != nil {
			logger.WithError(err).Error("Failed to retrieve tenant")
			continue
		}
		ctx := log.NewContextWithField(ctx, "tenant_id", tenant.TenantID)
		err = tbs.contactBackend(ctx, tenant, tenantData.Totals)
		if err != nil {
			logger.WithError(err).Error("Failed to report metrics to backend")
			continue
		}
	}
	return ttnpb.Empty, nil
}
