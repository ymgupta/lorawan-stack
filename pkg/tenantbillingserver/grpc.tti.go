// Copyright © 2019 The Things Industries B.V.

package tenantbillingserver

import (
	"context"

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Report implements ttipb.TbsServer.
func (tbs *TenantBillingServer) Report(ctx context.Context, data *ttipb.MeteringData) (*types.Empty, error) {
	// TODO: authenticate request.
	cc, err := tbs.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		return nil, err
	}
	client := ttipb.NewTenantRegistryClient(cc)
	for _, tenantData := range data.Tenants {
		tenant, err := client.Get(ctx, &ttipb.GetTenantRequest{
			TenantIdentifiers: tenantData.TenantIdentifiers,
			FieldMask: types.FieldMask{
				Paths: []string{
					"attributes",
				},
			},
		})
		if err != nil {
			return nil, err
		}
		for _, backend := range tbs.backends {
			err = backend.Report(ctx, tenant, tenantData.Totals)
			if err != nil {
				return nil, err
			}
		}
	}
	return ttnpb.Empty, nil
}
