// Copyright © 2019 The Things Industries B.V.

package middleware

import (
	"net/http"

	"go.thethings.network/lorawan-stack/v3/pkg/license"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
	"go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
)

// Middleware is HTTP middleware for extracting tenant IDs from the request.
func Middleware(config tenant.Config) webmiddleware.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if license.RequireMultiTenancy(ctx) != nil {
				// No multi-tenancy. All requests use the default tenant ID.
				if id := config.DefaultID; id != "" {
					ctx = tenant.NewContext(ctx, ttipb.TenantIdentifiers{TenantID: id})
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			if id := tenant.FromContext(ctx); !id.IsZero() {
				// Tenant already in context.
				if err := fetchTenant(ctx); err != nil {
					webhandlers.Error(w, r, err)
					return
				}
				next.ServeHTTP(w, r)
				return
			}
			if id := tenantID(r.URL.Hostname(), config); id != "" {
				// Derive tenant ID from hostname.
				ctx = tenant.NewContext(ctx, ttipb.TenantIdentifiers{TenantID: id})
				if err := fetchTenant(ctx); err != nil {
					webhandlers.Error(w, r, err)
					return
				}
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			if config.DefaultID == "" {
				// No default tenant ID either, so return "missing" error.
				webhandlers.Error(w, r, errMissingTenantID.New())
				return
			}
			webmiddleware.Redirect(webmiddleware.RedirectConfiguration{
				HostName: func(hostname string) string {
					for _, baseDomain := range config.BaseDomains {
						if hostname == baseDomain {
							return config.DefaultID + "." + hostname
						}
					}
					return hostname // No redirect, so ends up in the handler below.
				},
			})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				webhandlers.Error(w, r, errMissingTenantID.New())
				return
			})).ServeHTTP(w, r)
		})
	}
}
