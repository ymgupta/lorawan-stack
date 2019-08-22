// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deviceclaimingserver

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.thethings.network/lorawan-stack/pkg/component"
	web_errors "go.thethings.network/lorawan-stack/pkg/errors/web"
	"go.thethings.network/lorawan-stack/pkg/log"
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

// FrontendConfig is the configuration for the Device Claiming Server frontend.
type FrontendConfig struct {
	Language string          `json:"language" name:"-"`
	IS       webui.APIConfig `json:"is" name:"is"`
}

// Config is the configuration for the Device Claiming Server.
type Config struct {
	OAuth oauthclient.Config `name:"oauth"`
	Mount string             `name:"mount" description:"Path on the server where the Console will be served"`
	UI    UIConfig           `name:"ui"`
}

// DeviceClaimingServer is the Device Claiming Server.
type DeviceClaimingServer struct {
	*component.Component
	ctx    context.Context
	oc     *oauthclient.OAuthClient
	config Config

	grpc struct {
		endDeviceClaimingServer *endDeviceClaimingServer
	}
}

// New returns a new Device Claiming component.
func New(c *component.Component, config *Config) (*DeviceClaimingServer, error) {
	config.OAuth.StateCookieName = "_claim_state"
	config.OAuth.AuthCookieName = "_claim_auth"
	config.OAuth.RootURL = config.UI.CanonicalURL
	oc, err := oauthclient.New(c, config.OAuth)
	if err != nil {
		return nil, err
	}

	dcs := &DeviceClaimingServer{
		Component: c,
		oc:        oc,
		config:    *config,
		ctx:       log.NewContextWithField(c.Context(), "namespace", "deviceclaimingserver"),
	}

	dcs.grpc.endDeviceClaimingServer = &endDeviceClaimingServer{DCS: dcs}

	if dcs.config.Mount == "" {
		dcs.config.Mount = dcs.config.UI.MountPath()
	}

	c.RegisterGRPC(dcs)
	c.RegisterWeb(dcs)

	return dcs, nil
}

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

// RegisterRoutes implements web.Registerer. It registers the Console to the web server.
func (dcs *DeviceClaimingServer) RegisterRoutes(server *web.Server) {
	group := server.Group(
		dcs.config.Mount,
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				c.Set("template_data", dcs.config.UI.TemplateData)
				frontendConfig := dcs.config.UI.FrontendConfig
				frontendConfig.Language = dcs.config.UI.TemplateData.Language
				c.Set("app_config", struct {
					FrontendConfig
				}{
					FrontendConfig: frontendConfig.Apply(c.Request().Context()),
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
