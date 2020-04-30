// Copyright © 2019 The Things Industries B.V.

package middleware

import (
	"context"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/tenant"
	"google.golang.org/grpc/metadata"
)

func TestGRPCInterceptor(t *testing.T) {
	config := tenant.Config{
		DefaultID: "default",
		BaseDomains: []string{
			"nz.cluster.ttn",
			"identity.ttn",
		},
	}

	testCases := []struct {
		desc string
		ctx  context.Context
	}{
		// Set tenant ID (typically by forwarding auth)
		{desc: "tenant id in metadata", ctx: metadata.NewIncomingContext(context.Background(), metadata.MD{"tenant-id": []string{"foo-bar"}})},
		// Set authority (typically set by SDKs)
		{desc: "tenant id in authority", ctx: metadata.NewIncomingContext(context.Background(), metadata.MD{":authority": []string{"foo-bar"}})},
		// Set host name (typically set by gRPC clients)
		{desc: "host name in authority", ctx: metadata.NewIncomingContext(context.Background(), metadata.MD{":authority": []string{"foo-bar.identity.ttn"}})},
		// Set X-Forwarded-Host (typically set by gRPC gateway)
		{desc: "x-forwarded-host", ctx: metadata.NewIncomingContext(context.Background(), metadata.MD{"x-forwarded-host": []string{"foo-bar.nz.cluster.ttn"}})},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			assertions.New(t).So(fromRPCContext(tc.ctx, config).TenantID, should.Equal, "foo-bar")
		})
	}
}
