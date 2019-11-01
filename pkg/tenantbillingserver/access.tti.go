// Copyright © 2019 The Things Industries B.V.

package tenantbillingserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

var (
	errUnauthenticated = errors.DefineUnauthenticated("unauthenticated", "unauthenticated")
)

func (tbs *TenantBillingServer) billingRightsHook(h grpc.UnaryHandler) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		if p, ok := peer.FromContext(ctx); ok {
			for _, r := range tbs.config.reporterAddressRegexps {
				if r.MatchString(p.Addr.String()) {
					return h(ctx, req)
				}
			}
		}
		return nil, errUnauthenticated
	}
}
