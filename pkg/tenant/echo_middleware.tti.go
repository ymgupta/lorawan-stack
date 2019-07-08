// Copyright © 2019 The Things Industries B.V.

//+build tti

package tenant

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

func fromRequest(r *http.Request) (res ttipb.TenantIdentifiers) {
	if host := r.Header.Get("X-Forwarded-Host"); host != "" {
		res.TenantID = tenantID(host)
		return
	}
	if host := r.Host; host != "" {
		res.TenantID = tenantID(host)
		return
	}
	if tlsState := r.TLS; tlsState != nil {
		res.TenantID = tenantID(tlsState.ServerName)
		return
	}
	return
}

// Middleware is echo middleware for extracting tenant IDs from the request.
func Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		if id := FromContext(ctx); id.TenantID != "" {
			return next(c)
		}
		if id := fromRequest(c.Request()); id.TenantID != "" {
			c.SetRequest(c.Request().WithContext(NewContext(ctx, id)))
		} else if err := UseEmptyID(); err != nil {
			return err
		}
		return next(c)
	}
}
