// Copyright © 2019 The Things Industries B.V.

package deviceclaimingserver

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.thethings.network/lorawan-stack/pkg/component"
	web_errors "go.thethings.network/lorawan-stack/pkg/errors/web"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/web"
	"go.thethings.network/lorawan-stack/pkg/web/oauthclient"
	"go.thethings.network/lorawan-stack/pkg/webui"
	"google.golang.org/grpc"
)

// UIConfig is the combined configuration for the Device Claiming Server UI.
type UIConfig struct {
	webui.TemplateData `name:",squash"`
	FrontendConfig     `name:",squash"`
}

// StackConfig is the configuration of the stack components.
type StackConfig struct {
	IS  webui.APIConfig `json:"is" name:"is"`
	DCS webui.APIConfig `json:"dcs" name:"dcs"`
}

// FrontendConfig is the configuration for the Device Claiming Server frontend.
type FrontendConfig struct {
	Language    string `json:"language" name:"-"`
	StackConfig `json:"stack_config" name:",squash"`
}

// Config is the configuration for the Device Claiming Server.
type Config struct {
	OAuth                  oauthclient.Config            `name:"oauth"`
	Mount                  string                        `name:"mount" description:"Path on the server where the Console will be served"`
	UI                     UIConfig                      `name:"ui"`
	AuthorizedApplications AuthorizedApplicationRegistry `name:"-"`
}

// DeviceClaimingServer is the Device Claiming Server.
type DeviceClaimingServer struct {
	*component.Component
	ctx    context.Context
	oc     *oauthclient.OAuthClient
	config Config

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
	conf.OAuth.StateCookieName = "_claim_state"
	conf.OAuth.AuthCookieName = "_claim_auth"
	conf.OAuth.RootURL = conf.UI.CanonicalURL
	oc, err := oauthclient.New(c, conf.OAuth)
	if err != nil {
		return nil, err
	}

	dcs := &DeviceClaimingServer{
		Component:              c,
		ctx:                    log.NewContextWithField(c.Context(), "namespace", "deviceclaimingserver"),
		authorizedAppsRegistry: conf.AuthorizedApplications,
		oc:                     oc,
		config:                 *conf,
	}

	dcs.grpc.endDeviceClaimingServer = &endDeviceClaimingServer{DCS: dcs}

	for _, opt := range opts {
		opt(dcs)
	}

	if dcs.config.Mount == "" {
		dcs.config.Mount = dcs.config.UI.MountPath()
	}

	c.RegisterGRPC(dcs)
	c.RegisterWeb(dcs)

	return dcs, nil
}

type ctxKeyType struct{}

var ctxKey ctxKeyType

func (dcs *DeviceClaimingServer) configFromContext(ctx context.Context) *Config {
	if config, ok := ctx.Value(ctxKey).(*Config); ok {
		return config
	}
	config := dcs.config.Apply(ctx)
	return &config
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

// RegisterRoutes implements web.Registerer. It registers the Device Claiming Server to the web server.
func (dcs *DeviceClaimingServer) RegisterRoutes(server *web.Server) {
	group := server.Group(
		dcs.config.Mount,
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				config := dcs.configFromContext(c.Request().Context())
				c.Set("template_data", config.UI.TemplateData)
				frontendConfig := config.UI.FrontendConfig
				frontendConfig.Language = config.UI.TemplateData.Language
				c.Set("app_config", struct {
					FrontendConfig
				}{
					FrontendConfig: frontendConfig,
				})
				return next(c)
			}
		},
		web_errors.ErrorMiddleware(map[string]web_errors.ErrorRenderer{
			"text/html": webui.Template,
		}),
	)

	api := group.Group("/api", middleware.CSRF())
	api.GET("/auth/token", dcs.oc.HandleToken)
	api.POST("/auth/logout", dcs.oc.HandleLogout)

	page := group.Group("", middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:csrf",
	}))
	page.GET("/oauth/callback", dcs.oc.HandleCallback)

	group.GET("/login/ttn-stack", dcs.oc.HandleLogin)

	if dcs.config.Mount != "" && dcs.config.Mount != "/" {
		group.GET("", webui.Template.Handler, middleware.CSRF())
	}
	group.GET("/*", webui.Template.Handler, middleware.CSRF())
}

// WithTenantRegistry overrides the Device Claiming Server's tenant registry.
func WithTenantRegistry(registry ttipb.TenantRegistryClient) Option {
	return func(s *DeviceClaimingServer) {
		s.tenantRegistry = registry
	}
}

func (dcs *DeviceClaimingServer) getTenantRegistry(ctx context.Context) (ttipb.TenantRegistryClient, error) {
	if dcs.tenantRegistry != nil {
		return dcs.tenantRegistry, nil
	}
	conn, err := dcs.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
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

// Apply the context to the config.
func (conf Config) Apply(ctx context.Context) Config {
	deriv := conf
	deriv.OAuth = conf.OAuth.Apply(ctx)
	deriv.UI = conf.UI.Apply(ctx)
	return deriv
}

// Apply the context to the config.
func (conf UIConfig) Apply(ctx context.Context) UIConfig {
	deriv := conf
	deriv.TemplateData = conf.TemplateData.Apply(ctx)
	deriv.FrontendConfig = conf.FrontendConfig.Apply(ctx)
	return deriv
}

// Apply the context to the config.
func (conf StackConfig) Apply(ctx context.Context) StackConfig {
	deriv := conf
	deriv.IS = conf.IS.Apply(ctx)
	deriv.DCS = conf.DCS.Apply(ctx)
	return deriv
}

// Apply the context to the config.
func (conf FrontendConfig) Apply(ctx context.Context) FrontendConfig {
	deriv := conf
	deriv.StackConfig = conf.StackConfig.Apply(ctx)
	return deriv
}
