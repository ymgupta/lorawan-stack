// Copyright © 2019 The Things Industries B.V.

package tenant

import (
	"context"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func fromRPCContext(ctx context.Context) ttipb.TenantIdentifiers {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if id, ok := md["tenant-id"]; ok {
			return ttipb.TenantIdentifiers{TenantID: id[0]}
		}
		if host, ok := md["x-forwarded-host"]; ok { // Set by gRPC gateway.
			return ttipb.TenantIdentifiers{TenantID: tenantID(host[0])}
		}
		if authority, ok := md[":authority"]; ok { // Set by gRPC clients.
			return ttipb.TenantIdentifiers{TenantID: tenantID(authority[0])}
		}
	}
	return ttipb.TenantIdentifiers{}
}

// UnaryClientInterceptor is a gRPC interceptor that injects the tenant ID into the metadata.
func UnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if tenantID := FromContext(ctx); !tenantID.IsZero() {
		md, _ := metadata.FromOutgoingContext(ctx)
		ctx = metadata.NewOutgoingContext(ctx, metadata.Join(md, metadata.Pairs("tenant-id", tenantID.TenantID)))
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

// StreamClientInterceptor is a gRPC interceptor that injects the tenant ID into the metadata.
func StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	if tenantID := FromContext(ctx); !tenantID.IsZero() {
		md, _ := metadata.FromOutgoingContext(ctx)
		ctx = metadata.NewOutgoingContext(ctx, metadata.Join(md, metadata.Pairs("tenant-id", tenantID.TenantID)))
	}
	return streamer(ctx, desc, cc, method, opts...)
}

// UnaryServerInterceptor is a gRPC interceptor that extracts the tenant ID from the context.
func UnaryServerInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if id := FromContext(ctx); !id.IsZero() {
		return handler(ctx, req)
	}
	if id := fromRPCContext(ctx); !id.IsZero() {
		return handler(NewContext(ctx, id), req)
	}
	if err := UseEmptyID(); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

// StreamServerInterceptor is a gRPC interceptor that extracts the tenant ID from the context.
func StreamServerInterceptor(srv interface{}, stream grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := stream.Context()
	if id := FromContext(ctx); !id.IsZero() {
		return handler(srv, stream)
	}
	if id := fromRPCContext(ctx); !id.IsZero() {
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = NewContext(ctx, id)
		return handler(srv, wrapped)
	}
	if err := UseEmptyID(); err != nil {
		return err
	}
	return handler(srv, stream)
}
