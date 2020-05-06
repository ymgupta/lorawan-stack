// Copyright © 2019 The Things Industries B.V.

package tenantbillingserver_test

import (
	"context"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func mustHavePeer(ctx context.Context, c *component.Component, role ttnpb.ClusterRole) {
	for i := 0; i < 20; i++ {
		if _, err := c.GetPeer(ctx, role, nil); err == nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	panic("could not connect to peer")
}
