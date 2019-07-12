// Copyright © 2019 The Things Industries B.V.

package tenant

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
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
func Middleware(config Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			if id := FromContext(ctx); id.TenantID != "" {
				return next(c)
			}
			if id := config.DefaultID; id != "" {
				c.SetRequest(c.Request().WithContext(NewContext(ctx, ttipb.TenantIdentifiers{TenantID: id})))
				return next(c)
			}
			if id := fromRequest(c.Request()); id.TenantID != "" {
				c.SetRequest(c.Request().WithContext(NewContext(ctx, id)))
				return next(c)
			}
			return errMissingTenantID
		}
	}
}
