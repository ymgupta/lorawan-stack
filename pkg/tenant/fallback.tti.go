// Copyright © 2019 The Things Industries B.V.

//+build !tti

package tenant

import (
	"context"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"google.golang.org/grpc"
)

func FromContext(ctx context.Context) (res ttipb.TenantIdentifiers) {
	panic("wrong build tag for package tenant")
}

func NewContext(parent context.Context, id ttipb.TenantIdentifiers) context.Context {
	panic("wrong build tag for package tenant")
}

func Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	panic("wrong build tag for package tenant")
}

func UnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	panic("wrong build tag for package tenant")
}

func StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	panic("wrong build tag for package tenant")
}
