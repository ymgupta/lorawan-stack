// Copyright © 2019 The Things Industries B.V.

package tenantbillingserver

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/tenantbillingserver/stripe"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

// Config is the configuration for the TenantBillingServer.
type Config struct {
	Stripe stripe.Config `name:"stripe" description:"Stripe backend configuration"`
}

// TenantBillingServer is the Tenant Billing Server.
type TenantBillingServer struct {
	*component.Component
	ctx      context.Context
	backends []Backend
}

// Backend is an tenant handling backend.
type Backend interface {
	// Report reports the tenant registry totals to the backend. Only the attributes and IDs are retrieved for the tenant.
	Report(ctx context.Context, tenant *ttipb.Tenant, totals *ttipb.TenantRegistryTotals) error
}

// New returns a new Tenant Billing component.
func New(c *component.Component, conf *Config, opts ...Option) (*TenantBillingServer, error) {
	tbs := &TenantBillingServer{
		Component: c,
		ctx:       log.NewContextWithField(c.Context(), "namespace", "tenantbillingserver"),
	}

	for _, opt := range opts {
		opt(tbs)
	}

	if strp, err := conf.Stripe.New(c); err != nil {
		return nil, err
	} else if strp != nil {
		tbs.backends = append(tbs.backends, strp)
	}

	c.RegisterGRPC(tbs)
	return tbs, nil
}

// Option configures the TenantBillingServer.
type Option func(*TenantBillingServer)

// Context returns the context of the Device Claiming Server.
func (tbs *TenantBillingServer) Context() context.Context {
	return tbs.ctx
}

// Roles returns the roles that the Device Claiming Server fulfills.
func (tbs *TenantBillingServer) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_TENANT_BILLING_SERVER}
}

// RegisterServices registers services provided by dcs at s.
func (tbs *TenantBillingServer) RegisterServices(s *grpc.Server) {
	ttipb.RegisterTbsServer(s, tbs)
}

// RegisterHandlers registers gRPC handlers.
func (tbs *TenantBillingServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttipb.RegisterTbsHandler(tbs.Context(), s, conn)
}
