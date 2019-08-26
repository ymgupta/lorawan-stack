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

	grpc struct {
		endDeviceClaimingServer *endDeviceClaimingServer
	}
}

// New returns a new Device Claiming component.
func New(c *component.Component, conf *Config) (*DeviceClaimingServer, error) {
	dcs := &DeviceClaimingServer{
		Component:              c,
		ctx:                    log.NewContextWithField(c.Context(), "namespace", "deviceclaimingserver"),
		authorizedAppsRegistry: conf.AuthorizedApplications,
	}

	dcs.grpc.endDeviceClaimingServer = &endDeviceClaimingServer{DCS: dcs}

	c.RegisterGRPC(dcs)
	return dcs, nil
}

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

func (dcs *DeviceClaimingServer) getTenantRegistry(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) ttipb.TenantRegistryClient {
	if dcs.tenantRegistry != nil {
		return dcs.tenantRegistry
	}
	return ttipb.NewTenantRegistryClient(dcs.GetPeer(ctx, ttnpb.PeerInfo_ENTITY_REGISTRY, ids).Conn())
}
