// Copyright © 2019 The Things Industries B.V.

package middleware

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

func fromRequest(r *http.Request) ttipb.TenantIdentifiers {
	if host := r.Header.Get("X-Forwarded-Host"); host != "" {
		return ttipb.TenantIdentifiers{TenantID: tenantID(host)}
	}
	if host := r.Host; host != "" {
		return ttipb.TenantIdentifiers{TenantID: tenantID(host)}
	}
	if tlsState := r.TLS; tlsState != nil {
		return ttipb.TenantIdentifiers{TenantID: tenantID(tlsState.ServerName)}
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
					return next(c)
				}
				if id := fromRequest(c.Request()); id.TenantID != "" {
					c.SetRequest(c.Request().WithContext(tenant.NewContext(ctx, id)))
					return next(c)
				}
			}
			if id := config.DefaultID; id != "" {
				c.SetRequest(c.Request().WithContext(tenant.NewContext(ctx, ttipb.TenantIdentifiers{TenantID: id})))
				return next(c)
			}
			return errMissingTenantID
		}
	}
}
