// Copyright © 2019 The Things Industries B.V.

package middleware

import (
	"context"
	"fmt"
	"time"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/warning"
	"google.golang.org/grpc"
)

func checkLicense(ctx context.Context) error {
	l := license.FromContext(ctx)
	if err := license.CheckValidity(&l); err != nil {
		return err
	}
	if validUntil := l.GetValidUntil(); !validUntil.IsZero() && time.Until(validUntil) < l.GetWarnFor() {
		if l.Metering != nil {
			warning.Add(ctx, fmt.Sprintf("failed to report to metering service, license expiry at %s", validUntil))
		} else {
			warning.Add(ctx, fmt.Sprintf("license expiry at %s", validUntil))
		}
	}
	return nil
}

// Middleware is an Echo middleware verifying the license validity on each request.
func Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := checkLicense(c.Request().Context()); err != nil {
			return err
		}
		return next(c)
	}
}

// UnaryServerInterceptor is a gRPC interceptor verifying the license validity on each request.
func UnaryServerInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if err := checkLicense(ctx); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

// StreamServerInterceptor is a gRPC interceptor verifying the license validity on each request.
func StreamServerInterceptor(srv interface{}, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := checkLicense(stream.Context()); err != nil {
		return err
	}
	return handler(srv, stream)
}
