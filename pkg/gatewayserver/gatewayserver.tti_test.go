// Copyright © 2019 The Things Industries B.V.

//+build tti

package gatewayserver_test

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/gatewayserver"
	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

func init() {
	gatewayserver.CustomContextFromIdentifier = func(ctx context.Context, ids ttnpb.GatewayIdentifiers) (context.Context, error) {
		if tenant.FromContext(ctx).TenantID == "" {
			return tenant.NewContext(ctx, tenant.FromContext(test.Context())), nil
		}
		return ctx, nil
	}
}
