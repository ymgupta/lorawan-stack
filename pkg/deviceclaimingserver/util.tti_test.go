// Copyright © 2019 The Things Industries B.V.

package deviceclaimingserver_test

import (
	"context"
	"net"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

func mustHavePeer(ctx context.Context, c *component.Component, role ttnpb.ClusterRole) {
	for i := 0; i < 20; i++ {
		time.Sleep(20 * time.Millisecond)
		if _, err := c.GetPeer(ctx, role, nil); err == nil {
			return
		}
	}
	panic("could not connect to peer")
}

type mockAuthorizedApplicationsRegistry struct {
	GetFunc func(context.Context, ttnpb.ApplicationIdentifiers, []string) (*ttipb.ApplicationAPIKey, error)
	SetFunc func(context.Context, ttnpb.ApplicationIdentifiers, []string, func(*ttipb.ApplicationAPIKey) (*ttipb.ApplicationAPIKey, []string, error)) (*ttipb.ApplicationAPIKey, error)
}

func (r *mockAuthorizedApplicationsRegistry) Get(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttipb.ApplicationAPIKey, error) {
	if r.GetFunc == nil {
		panic("Get called, but not set")
	}
	return r.GetFunc(ctx, ids, paths)
}

func (r *mockAuthorizedApplicationsRegistry) Set(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string, f func(*ttipb.ApplicationAPIKey) (*ttipb.ApplicationAPIKey, []string, error)) (*ttipb.ApplicationAPIKey, error) {
	if r.SetFunc == nil {
		panic("Set called, but not set")
	}
	return r.SetFunc(ctx, ids, paths, f)
}

type mockTenantRegistry struct {
	ttipb.TenantRegistryClient
	GetIdentifiersForEndDeviceEUIsFunc func(context.Context, *ttipb.GetTenantIdentifiersForEndDeviceEUIsRequest, ...grpc.CallOption) (*ttipb.TenantIdentifiers, error)
}

func (r *mockTenantRegistry) GetIdentifiersForEndDeviceEUIs(ctx context.Context, in *ttipb.GetTenantIdentifiersForEndDeviceEUIsRequest, opts ...grpc.CallOption) (*ttipb.TenantIdentifiers, error) {
	if r.GetIdentifiersForEndDeviceEUIsFunc == nil {
		panic("GetIdentifiersForEndDeviceEUIs called, but not set")
	}
	return r.GetIdentifiersForEndDeviceEUIsFunc(ctx, in, opts...)
}

type mockApplicationAccess struct {
	ttnpb.ApplicationAccessClient
	ListRightsFunc func(context.Context, *ttnpb.ApplicationIdentifiers, ...grpc.CallOption) (*ttnpb.Rights, error)
}

func (c *mockApplicationAccess) ListRights(ctx context.Context, in *ttnpb.ApplicationIdentifiers, opts ...grpc.CallOption) (*ttnpb.Rights, error) {
	if c.ListRightsFunc == nil {
		panic("ListRightsFunc called, but not set")
	}
	return c.ListRightsFunc(ctx, in, opts...)
}

type mockDeviceRegistry struct {
	ttnpb.EndDeviceRegistryClient
	CreateFunc                func(context.Context, *ttnpb.CreateEndDeviceRequest, ...grpc.CallOption) (*ttnpb.EndDevice, error)
	GetFunc                   func(context.Context, *ttnpb.GetEndDeviceRequest, ...grpc.CallOption) (*ttnpb.EndDevice, error)
	GetIdentifiersForEUIsFunc func(context.Context, *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error)
	DeleteFunc                func(context.Context, *ttnpb.EndDeviceIdentifiers, ...grpc.CallOption) (*pbtypes.Empty, error)
}

func (r *mockDeviceRegistry) Create(ctx context.Context, in *ttnpb.CreateEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
	if r.CreateFunc == nil {
		panic("Create called, but not set")
	}
	return r.CreateFunc(ctx, in, opts...)
}

func (r *mockDeviceRegistry) Get(ctx context.Context, in *ttnpb.GetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
	if r.GetFunc == nil {
		panic("Get called, but not set")
	}
	return r.GetFunc(ctx, in, opts...)
}

func (r *mockDeviceRegistry) GetIdentifiersForEUIs(ctx context.Context, in *ttnpb.GetEndDeviceIdentifiersForEUIsRequest, opts ...grpc.CallOption) (*ttnpb.EndDeviceIdentifiers, error) {
	if r.GetIdentifiersForEUIsFunc == nil {
		panic("GetIdentifiersForEUIs called, but not set")
	}
	return r.GetIdentifiersForEUIsFunc(ctx, in, opts...)
}

func (r *mockDeviceRegistry) Delete(ctx context.Context, in *ttnpb.EndDeviceIdentifiers, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
	if r.DeleteFunc == nil {
		panic("Delete called, but not set")
	}
	return r.DeleteFunc(ctx, in, opts...)
}

type mockJsDeviceRegistry struct {
	ttnpb.JsEndDeviceRegistryClient
	GetFunc    func(context.Context, *ttnpb.GetEndDeviceRequest, ...grpc.CallOption) (*ttnpb.EndDevice, error)
	SetFunc    func(context.Context, *ttnpb.SetEndDeviceRequest, ...grpc.CallOption) (*ttnpb.EndDevice, error)
	DeleteFunc func(context.Context, *ttnpb.EndDeviceIdentifiers, ...grpc.CallOption) (*pbtypes.Empty, error)
}

func (r *mockJsDeviceRegistry) Get(ctx context.Context, in *ttnpb.GetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
	if r.GetFunc == nil {
		panic("Get called, but not set")
	}
	return r.GetFunc(ctx, in, opts...)
}

func (r *mockJsDeviceRegistry) Set(ctx context.Context, in *ttnpb.SetEndDeviceRequest, opts ...grpc.CallOption) (*ttnpb.EndDevice, error) {
	if r.SetFunc == nil {
		panic("Set called, but not set")
	}
	return r.SetFunc(ctx, in, opts...)
}

func (r *mockJsDeviceRegistry) Delete(ctx context.Context, in *ttnpb.EndDeviceIdentifiers, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
	if r.DeleteFunc == nil {
		panic("Delete called, but not set")
	}
	return r.DeleteFunc(ctx, in, opts...)
}

func startMockNS(ctx context.Context, opts ...rpcserver.Option) (*mockNS, string) {
	ns := &mockNS{}
	srv := rpcserver.New(ctx, opts...)
	ttnpb.RegisterNsEndDeviceRegistryServer(srv.Server, ns)
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	go srv.Serve(lis)
	go func() {
		<-ctx.Done()
		lis.Close()
	}()
	return ns, lis.Addr().String()
}

type mockNS struct {
	GetFunc    func(context.Context, *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error)
	SetFunc    func(context.Context, *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error)
	DeleteFunc func(context.Context, *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error)
}

func (ns *mockNS) Get(ctx context.Context, in *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if ns.GetFunc == nil {
		panic("Get called, but not set")
	}
	return ns.GetFunc(ctx, in)
}

func (ns *mockNS) Set(ctx context.Context, in *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if ns.SetFunc == nil {
		panic("Set called, but not set")
	}
	return ns.SetFunc(ctx, in)
}

func (ns *mockNS) Delete(ctx context.Context, in *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	if ns.DeleteFunc == nil {
		panic("Delete called, but not set")
	}
	return ns.DeleteFunc(ctx, in)
}

func startMockAS(ctx context.Context, opts ...rpcserver.Option) (*mockAS, string) {
	as := &mockAS{}
	srv := rpcserver.New(ctx, opts...)
	ttnpb.RegisterAsEndDeviceRegistryServer(srv.Server, as)
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	go srv.Serve(lis)
	go func() {
		<-ctx.Done()
		lis.Close()
	}()
	return as, lis.Addr().String()
}

type mockAS struct {
	GetFunc    func(context.Context, *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error)
	SetFunc    func(context.Context, *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error)
	DeleteFunc func(context.Context, *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error)
}

func (as *mockAS) Get(ctx context.Context, in *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if as.GetFunc == nil {
		panic("Get called, but not set")
	}
	return as.GetFunc(ctx, in)
}

func (as *mockAS) Set(ctx context.Context, in *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if as.SetFunc == nil {
		panic("Set called, but not set")
	}
	return as.SetFunc(ctx, in)
}

func (as *mockAS) Delete(ctx context.Context, in *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	if as.DeleteFunc == nil {
		panic("Delete called, but not set")
	}
	return as.DeleteFunc(ctx, in)
}
