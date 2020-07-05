// Copyright © 2019 The Things Industries B.V.

package deviceclaimingserver

import (
	"context"
	"strings"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func dialTenantCluster(ctx context.Context, role ttnpb.ClusterRole, target string, _ credentials.TransportCredentials, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	parts := strings.SplitN(target, ".", 2)
	if len(parts) == 2 {
		target = parts[1]
	}
	return grpc.DialContext(ctx, target, append(opts, grpc.WithInsecure())...)
}

func init() {
	customSameHost = strings.EqualFold
	customDialer = dialTenantCluster
	dialTimeout = 1 * time.Second
}
