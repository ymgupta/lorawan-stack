// Copyright © 2019 The Things Industries B.V.

package tenantbillingserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

var (
	errPeerAddressNotAllowed = errors.DefineUnauthenticated("peer_address_not_allowed", "peer address `{peer_address}` is not allowed")
)

func (tbs *TenantBillingServer) billingRightsUnaryHook(h grpc.UnaryHandler) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		p, ok := peer.FromContext(ctx)
		if !ok {
			panic("Peer missing from context")
		}
		for _, r := range tbs.config.reporterAddressRegexps {
			if r.MatchString(p.Addr.String()) {
				return h(ctx, req)
			}
		}
		return nil, errPeerAddressNotAllowed.WithAttributes("peer_address", p.Addr.String())
	}
}
