// Copyright © 2019 The Things Industries B.V.

package tbsmetrics

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

// ConnectionProvider is the interface used for connection.
type ConnectionProvider interface {
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole) (*grpc.ClientConn, error)
}

// Reporter is a license.MeteringReporter that reports the stats
// to the Tenant Billing Server found in the cluster.
type Reporter struct {
	connProvider ConnectionProvider
}

// New returns a new license.MeteringReporter that reports the metrics to the Tenant Billing Server of the cluster.
func New(config *ttipb.MeteringConfiguration_TenantBillingServer, connProvider ConnectionProvider) (*Reporter, error) {
	return &Reporter{
		connProvider: connProvider,
	}, nil
}

// Report implements license.MeteringReporter.
func (r *Reporter) Report(ctx context.Context, data *ttipb.MeteringData) error {
	cc, err := r.connProvider.GetPeerConn(ctx, ttnpb.ClusterRole_TENANT_BILLING_SERVER)
	if err != nil {
		return err
	}
	client := ttipb.NewTbsClient(cc)
	_, err = client.Report(ctx, data)
	if err != nil {
		return err
	}
	return nil
}
