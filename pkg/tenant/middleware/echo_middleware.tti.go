// Copyright © 2019 The Things Industries B.V.

package middleware

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

func fromRequest(r *http.Request, config tenant.Config) ttipb.TenantIdentifiers {
	if host := r.Header.Get("X-Forwarded-Host"); host != "" {
		return ttipb.TenantIdentifiers{TenantID: tenantID(host, config)}
	}
	if host := r.Host; host != "" {
		return ttipb.TenantIdentifiers{TenantID: tenantID(host, config)}
	}
	if tlsState := r.TLS; tlsState != nil {
		return ttipb.TenantIdentifiers{TenantID: tenantID(tlsState.ServerName, config)}
	}
	return ttipb.TenantIdentifiers{}
}

// Middleware is echo middleware for extracting tenant IDs from the request.
func Middleware(config tenant.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			if license.RequireMultiTenancy(ctx) == nil {
				if id := tenant.FromContext(ctx); id.TenantID != "" {
					if err := fetchTenant(ctx); err != nil {
						return err
					}
					return next(c)
				}
				if id := fromRequest(c.Request(), config); id.TenantID != "" {
					ctx = tenant.NewContext(ctx, id)
					c.SetRequest(c.Request().WithContext(ctx))
					if err := fetchTenant(ctx); err != nil {
						return err
					}
					return next(c)
				}
			}
			if id := config.DefaultID; id != "" {
				ctx = tenant.NewContext(ctx, ttipb.TenantIdentifiers{TenantID: id})
				c.SetRequest(c.Request().WithContext(ctx))
				return next(c)
			}
			return errMissingTenantID.New()
		}
	}
}
