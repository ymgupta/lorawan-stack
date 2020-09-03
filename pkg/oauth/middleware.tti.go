// Copyright © 2020 The Things Industries B.V.

package oauth

import (
	"github.com/gogo/protobuf/types"
	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
)

func (s *server) withFederatedProviders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		providers, err := s.store.FindAuthenticationProviders(ctx, nil, &types.FieldMask{Paths: []string{"ids", "name"}})
		if err != nil {
			return err
		}
		c.Set("page_data", struct {
			Providers []*ttipb.AuthenticationProvider `json:"providers"`
		}{
			Providers: providers,
		})
		return next(c)
	}
}
