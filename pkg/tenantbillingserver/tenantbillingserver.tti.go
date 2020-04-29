// Copyright © 2020 The Things Industries B.V.

package tenantbillingserver

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/tenantbillingserver/backend"
	"go.thethings.network/lorawan-stack/pkg/tenantbillingserver/stripe"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

// TenantBillingServer is the Tenant Billing Server.
type TenantBillingServer struct {
	*component.Component
	ctx context.Context

	config *Config

	backends struct {
		stripe backend.Interface
	}
}

const (
	tenantAdminAuthType = "TenantAdminKey"
)

// New returns a new Tenant Billing component.
func New(c *component.Component, conf *Config, opts ...Option) (*TenantBillingServer, error) {
	if err := license.RequireComponent(c.Context(), ttnpb.ClusterRole_TENANT_BILLING_SERVER); err != nil {
		return nil, err
	}

	tbs := &TenantBillingServer{
		Component: c,
		ctx:       log.NewContextWithField(c.Context(), "namespace", "tenantbillingserver"),
		config:    conf,
	}

	for _, opt := range opts {
		opt(tbs)
	}

	if err := tbs.config.compileRegex(tbs.ctx); err != nil {
		return nil, err
	}

	if tbs.config.PullInterval != 0 {
		c.RegisterTask(tbs.ctx, "collect_metering_data", tbs.run, component.TaskRestartOnFailure)
	}

	tenantAuth := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:  tenantAdminAuthType,
		AuthValue: conf.TenantAdminKey,
	})

	if conf.Stripe.Enable {
		strp, err := conf.Stripe.New(tbs.ctx, c, stripe.WithTenantRegistryAuth(tenantAuth))
		if err != nil {
			return nil, err
		} else if strp != nil {
			tbs.backends.stripe = strp
			c.RegisterWeb(strp)
		}
	}

	hooks.RegisterUnaryHook("/tti.lorawan.v3.Tbs", "billing-rights", tbs.billingRightsUnaryHook)

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
