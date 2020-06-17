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

func init() {
	customSameHost = strings.EqualFold
	customDialer = func(ctx context.Context, role ttnpb.ClusterRole, target string, _ credentials.TransportCredentials, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return grpc.DialContext(ctx, target, append(opts, grpc.WithInsecure())...)
	}
	dialTimeout = 1 * time.Second
}
