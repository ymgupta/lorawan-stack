// Copyright © 2019 The Things Industries B.V.

package oauth

import (
	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

var errInvalidProvider = errors.DefineInvalidArgument("invalid_provider", "the provider `{provider_id}` is invalid")

const oidcProviderID = "oidc"

func (s *server) FederatedLogin(c echo.Context) error {
	providerID := c.Param("provider")
	switch providerID {
	case oidcProviderID:
		return s.oidc.Login(c)
	}
	return errInvalidProvider.WithAttributes("provider_id", providerID)
}

func (s *server) FederatedCallback(c echo.Context) error {
	providerID := c.Param("provider")
	switch providerID {
	case oidcProviderID:
		return s.oidc.Callback(c)
	}
	return errInvalidProvider.WithAttributes("provider_id", providerID)
}
