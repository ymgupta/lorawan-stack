// Copyright © 2019 The Things Industries B.V.

package deviceclaimingserver

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

// Config is the configuration for the Device Claiming Server.
type Config struct {
	AuthorizedApplications AuthorizedApplicationRegistry `name:"-"`
}

// DeviceClaimingServer is the Device Claiming Server.
type DeviceClaimingServer struct {
	*component.Component
	ctx context.Context

	authorizedAppsRegistry AuthorizedApplicationRegistry
	tenantRegistry         ttipb.TenantRegistryClient
	applicationAccess      ttnpb.ApplicationAccessClient
	deviceRegistry         ttnpb.EndDeviceRegistryClient
	jsDeviceRegistry       ttnpb.JsEndDeviceRegistryClient

	grpc struct {
		endDeviceClaimingServer *endDeviceClaimingServer
	}
}

// New returns a new Device Claiming component.
func New(c *component.Component, conf *Config, opts ...Option) (*DeviceClaimingServer, error) {
	dcs := &DeviceClaimingServer{
		Component:              c,
		ctx:                    log.NewContextWithField(c.Context(), "namespace", "deviceclaimingserver"),
		authorizedAppsRegistry: conf.AuthorizedApplications,
	}

	dcs.grpc.endDeviceClaimingServer = &endDeviceClaimingServer{DCS: dcs}

	for _, opt := range opts {
		opt(dcs)
	}
	c.RegisterGRPC(dcs)
	return dcs, nil
}

// Option configures the DeviceClaimingServer.
type Option func(*DeviceClaimingServer)

// Context returns the context of the Device Claiming Server.
func (dcs *DeviceClaimingServer) Context() context.Context {
	return dcs.ctx
}

// Roles returns the roles that the Device Claiming Server fulfills.
func (dcs *DeviceClaimingServer) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_DEVICE_CLAIMING_SERVER}
}

// RegisterServices registers services provided by dcs at s.
func (dcs *DeviceClaimingServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterEndDeviceClaimingServerServer(s, dcs.grpc.endDeviceClaimingServer)
}

// RegisterHandlers registers gRPC handlers.
func (dcs *DeviceClaimingServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterEndDeviceClaimingServerHandler(dcs.Context(), s, conn)
}

// WithTenantRegistry overrides the Device Claiming Server's tenant registry.
func WithTenantRegistry(registry ttipb.TenantRegistryClient) Option {
	return func(s *DeviceClaimingServer) {
		s.tenantRegistry = registry
	}
}

func (dcs *DeviceClaimingServer) getTenantRegistry(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (ttipb.TenantRegistryClient, error) {
	if dcs.tenantRegistry != nil {
		return dcs.tenantRegistry, nil
	}
	conn, err := dcs.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, ids)
	if err != nil {
		return nil, err
	}
	return ttipb.NewTenantRegistryClient(conn), nil
}

// WithApplicationAccess overrides the Device Claiming Server's application access provider.
func WithApplicationAccess(access ttnpb.ApplicationAccessClient) Option {
	return func(s *DeviceClaimingServer) {
		s.applicationAccess = access
	}
}

func (dcs *DeviceClaimingServer) getApplicationAccess(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (ttnpb.ApplicationAccessClient, error) {
	if dcs.applicationAccess != nil {
		return dcs.applicationAccess, nil
	}
	conn, err := dcs.GetPeerConn(ctx, ttnpb.ClusterRole_ACCESS, ids)
	if err != nil {
		return nil, err
	}
	return ttnpb.NewApplicationAccessClient(conn), nil
}

// WithDeviceRegistry overrides the Device Claiming Server's Entity Registry device registry.
func WithDeviceRegistry(registry ttnpb.EndDeviceRegistryClient) Option {
	return func(s *DeviceClaimingServer) {
		s.deviceRegistry = registry
	}
}

func (dcs *DeviceClaimingServer) getDeviceRegistry(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (ttnpb.EndDeviceRegistryClient, error) {
	if dcs.deviceRegistry != nil {
		return dcs.deviceRegistry, nil
	}
	conn, err := dcs.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, ids)
	if err != nil {
		return nil, err
	}
	return ttnpb.NewEndDeviceRegistryClient(conn), nil
}

// WithJsDeviceRegistry overrides the Device Claiming Server's Join Server device registry.
func WithJsDeviceRegistry(registry ttnpb.JsEndDeviceRegistryClient) Option {
	return func(s *DeviceClaimingServer) {
		s.jsDeviceRegistry = registry
	}
}

func (dcs *DeviceClaimingServer) getJsDeviceRegistry(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (ttnpb.JsEndDeviceRegistryClient, error) {
	if dcs.jsDeviceRegistry != nil {
		return dcs.jsDeviceRegistry, nil
	}
	conn, err := dcs.GetPeerConn(ctx, ttnpb.ClusterRole_JOIN_SERVER, ids)
	if err != nil {
		return nil, err
	}
	return ttnpb.NewJsEndDeviceRegistryClient(conn), nil
}
