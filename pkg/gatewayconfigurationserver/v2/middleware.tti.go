// Copyright © 2020 The Things Industries B.V.

package gatewayconfigurationserver

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
	"go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
)

// validateAndFillIDsMultiTenant validates and fills the gateway ID in a tenant-aware manner.
// Since the GCS V2 endpoint is not using tenant middleware, the tenant is taken from the gateway UID if the environment
// is multi-tenant. Otherwise, the default tenant ID is used.
func validateAndFillIDsMultiTenant(config tenant.Config) webmiddleware.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			givenID := mux.Vars(r)["gateway_id_or_uid"]
			ids, err := unique.ToGatewayID(givenID)
			if err != nil {
				if config.DefaultID != "" {
					ctx = tenant.NewContext(ctx, ttipb.TenantIdentifiers{TenantID: config.DefaultID})
				} else {
					webhandlers.Error(w, r, err)
					return
				}
				ids = ttnpb.GatewayIdentifiers{
					GatewayID: givenID,
				}
			} else if ctx, err = unique.WithContext(ctx, givenID); err != nil {
				webhandlers.Error(w, r, err)
				return
			}
			if err := ids.ValidateContext(ctx); err != nil {
				webhandlers.Error(w, r, err)
				return
			}
			ctx = withGatewayID(ctx, ids)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
