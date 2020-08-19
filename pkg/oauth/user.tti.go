// Copyright © 2019 The Things Industries B.V.

package oauth

import (
	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/oauth/oidc"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
)

// FederatedAuthenticationProvider implements the authentication callbacks for
// federated authentication.
type FederatedAuthenticationProvider interface {
	Login(echo.Context) error
	Callback(echo.Context) error
}

var errInvalidProvider = errors.DefineInvalidArgument("invalid_provider", "the provider `{provider_id}` is invalid")

const oidcProviderID = "oidc"

func (s *server) routeFederatedRequest(c echo.Context, f func(FederatedAuthenticationProvider) error) error {
	ctx := c.Request().Context()
	providerID := c.Param("provider")
	provider, err := s.store.GetAuthenticationProvider(ctx, &ttipb.AuthenticationProviderIdentifiers{
		ProviderID: providerID,
	}, nil)
	if err != nil {
		return errInvalidProvider.WithCause(err).WithAttributes("provider_id", providerID)
	}
	switch provider.Configuration.Provider.(type) {
	case *ttipb.AuthenticationProvider_Configuration_OIDC:
		{
			ctx, err := oidc.WithProvider(ctx, provider)
			if err != nil {
				return err
			}
			c.SetRequest(c.Request().WithContext(ctx))
			return f(s.providers.oidc)
		}
	default:
		panic("unknown authentication provider type")
	}
}

func (s *server) FederatedLogin(c echo.Context) error {
	return s.routeFederatedRequest(c, func(p FederatedAuthenticationProvider) error { return p.Login(c) })
}

func (s *server) FederatedCallback(c echo.Context) error {
	return s.routeFederatedRequest(c, func(p FederatedAuthenticationProvider) error { return p.Callback(c) })
}
