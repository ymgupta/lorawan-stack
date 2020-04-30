// Copyright © 2019 The Things Industries B.V.

package middleware

import (
	"context"
	"strings"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func fromRPCContext(ctx context.Context, config tenant.Config) ttipb.TenantIdentifiers {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if id, ok := md["tenant-id"]; ok {
			return ttipb.TenantIdentifiers{TenantID: id[0]}
		}
		if host, ok := md["x-forwarded-host"]; ok { // Set by gRPC gateway.
			return ttipb.TenantIdentifiers{TenantID: tenantID(host[0], config)}
		}
		if authority, ok := md[":authority"]; ok { // Set by gRPC clients.
			if authority[0] != "in-process" {
				return ttipb.TenantIdentifiers{TenantID: tenantID(authority[0], config)}
			}
		}
	}
	return ttipb.TenantIdentifiers{}
}

// UnaryClientInterceptor is a gRPC interceptor that injects the tenant ID into the metadata.
func UnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if license.RequireMultiTenancy(ctx) == nil {
		if tenantID := tenant.FromContext(ctx); !tenantID.IsZero() {
			md, _ := metadata.FromOutgoingContext(ctx)
			ctx = metadata.NewOutgoingContext(ctx, metadata.Join(md, metadata.Pairs("tenant-id", tenantID.TenantID)))
		}
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

// StreamClientInterceptor is a gRPC interceptor that injects the tenant ID into the metadata.
func StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	if license.RequireMultiTenancy(ctx) == nil {
		if tenantID := tenant.FromContext(ctx); !tenantID.IsZero() {
			md, _ := metadata.FromOutgoingContext(ctx)
			ctx = metadata.NewOutgoingContext(ctx, metadata.Join(md, metadata.Pairs("tenant-id", tenantID.TenantID)))
		}
	}
	return streamer(ctx, desc, cc, method, opts...)
}

var tenantAgnosticServices = []string{"/tti.lorawan.v3.TenantRegistry", "/tti.lorawan.v3.Tbs"}

// UnaryServerInterceptor is a gRPC interceptor that extracts the tenant ID from the context.
func UnaryServerInterceptor(config tenant.Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if license.RequireMultiTenancy(ctx) == nil {
			if id := tenant.FromContext(ctx); !id.IsZero() {
				if err := fetchTenant(ctx); err != nil {
					return nil, err
				}
				return handler(ctx, req)
			}
			if id := fromRPCContext(ctx, config); !id.IsZero() {
				ctx = tenant.NewContext(ctx, id)
				if err := fetchTenant(ctx); err != nil {
					return nil, err
				}
				return handler(ctx, req)
			}
		}
		if id := config.DefaultID; id != "" {
			ctx = tenant.NewContext(ctx, ttipb.TenantIdentifiers{TenantID: id})
			return handler(ctx, req)
		}
		for _, service := range tenantAgnosticServices {
			if strings.HasPrefix(info.FullMethod, service) {
				return handler(ctx, req)
			}
		}
		return nil, errMissingTenantID.New()
	}
}

// StreamServerInterceptor is a gRPC interceptor that extracts the tenant ID from the context.
func StreamServerInterceptor(config tenant.Config) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		if license.RequireMultiTenancy(ctx) == nil {
			if id := tenant.FromContext(ctx); !id.IsZero() {
				if err := fetchTenant(ctx); err != nil {
					return err
				}
				return handler(srv, stream)
			}
			if id := fromRPCContext(ctx, config); !id.IsZero() {
				ctx = tenant.NewContext(ctx, id)
				wrapped := grpc_middleware.WrapServerStream(stream)
				wrapped.WrappedContext = ctx
				if err := fetchTenant(ctx); err != nil {
					return err
				}
				return handler(srv, wrapped)
			}
		}
		if id := config.DefaultID; id != "" {
			ctx = tenant.NewContext(ctx, ttipb.TenantIdentifiers{TenantID: id})
			wrapped := grpc_middleware.WrapServerStream(stream)
			wrapped.WrappedContext = ctx
			return handler(srv, wrapped)
		}
		for _, service := range tenantAgnosticServices {
			if strings.HasPrefix(info.FullMethod, service) {
				return handler(srv, stream)
			}
		}
		return errMissingTenantID.New()
	}
}
