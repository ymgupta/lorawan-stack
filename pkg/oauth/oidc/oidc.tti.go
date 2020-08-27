// Copyright © 2020 The Things Industries B.V.

package oidc

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc"
	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/web/oauthclient"
	"golang.org/x/oauth2"
)

// Server is an OIDC federated authentication Server.
type Server struct {
	UpstreamServer
	store Store
	oc    *oauthclient.OAuthClient
}

// New creates a new OIDC federated authentication provider wraping the provided upstream.
func New(c *component.Component, upstream UpstreamServer, store Store) (*Server, error) {
	s := &Server{
		UpstreamServer: upstream,
		store:          store,
	}
	var err error
	s.oc, err = oauthclient.New(c,
		oauthclient.Config{
			RootURL:         upstream.GetTemplateData(c.Context()).CanonicalURL,
			StateCookieName: "_oidc_state",
			AuthCookieName:  "_oidc_auth",
		},
		oauthclient.WithNextKey("n"),
		oauthclient.WithOAuth2ConfigProvider(s.oauth),
		oauthclient.WithCallback(s.callback),
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Server) oidc(ctx context.Context) (*oauth2.Config, *oidc.Provider, error) {
	config := configFromContext(ctx)
	provider, err := oidc.NewProvider(ctx, config.ProviderURL)
	if err != nil {
		return nil, nil, err
	}
	uiConfig := s.GetTemplateData(ctx)
	redirectURL := fmt.Sprintf("%s/login/%s/callback", strings.TrimSuffix(uiConfig.CanonicalURL, "/"), config.ID)
	return &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}, provider, nil
}

func (s *Server) oauth(c echo.Context) *oauth2.Config {
	ctx := c.Request().Context()
	conf, _, err := s.oidc(ctx)
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to retrieve OAuth2 configuration")
		return nil
	}
	return conf
}

var (
	errInvalidState = errors.DefineInvalidArgument("invalid_state", "invalid state `{state}` provided")
	errNoIDToken    = errors.DefineNotFound("no_id_token", "no ID token found")
)

func (s *Server) callback(c echo.Context, token *oauth2.Token, next string) error {
	ctx := c.Request().Context()
	config, provider, err := s.oidc(ctx)
	if err != nil {
		return err
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return errNoIDToken.New()
	}
	verifier := provider.Verifier(&oidc.Config{
		ClientID: config.ClientID,
	})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return err
	}
	var claims claimsToken
	if err := idToken.Claims(&claims); err != nil {
		return err
	}
	oidcConfig := configFromContext(ctx)
	providerIDs := &ttipb.AuthenticationProviderIdentifiers{ProviderID: oidcConfig.ID}
	externalUser, err := s.store.GetExternalUserByExternalID(ctx, providerIDs, claims.Subject)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if errors.IsNotFound(err) {
		externalUser, err = s.createExternalUser(ctx, &claims)
		if err != nil {
			return err
		}
	}
	if err := s.CreateUserSession(c, externalUser.UserIDs); err != nil {
		return err
	}
	url, err := url.Parse(next)
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, fmt.Sprintf("%s?%s", url.Path, url.RawQuery))
}

var (
	errRegistrationsDisabled = errors.DefineFailedPrecondition("registrations_disabled", "account registrations are disabled")
	errNoUsernameAvailable   = errors.DefineResourceExhausted("no_username_available", "no username available")
)

func (s *Server) createExternalUser(ctx context.Context, claims *claimsToken) (*ttipb.ExternalUser, error) {
	config := configFromContext(ctx)
	if !config.AllowRegistrations {
		return nil, errRegistrationsDisabled.New()
	}
	now := time.Now()
	password, err := auth.Hash(ctx, random.String(64))
	if err != nil {
		return nil, err
	}
	newUser := &ttnpb.User{
		Name:                claims.Name,
		PrimaryEmailAddress: claims.Email,
		Password:            password,
		PasswordUpdatedAt:   &now,
		State:               ttnpb.STATE_APPROVED,
	}
	if claims.EmailVerified {
		newUser.PrimaryEmailAddressValidatedAt = &now
	}
	var ids ttnpb.UserIdentifiers
	var found bool
	for _, username := range usernames(config.ID, claims) {
		ids = ttnpb.UserIdentifiers{UserID: username}
		newUser.UserIdentifiers = ids
		if _, err := s.store.CreateUser(ctx, newUser); err == nil {
			found = true
			break
		}
	}
	if !found {
		return nil, errNoUsernameAvailable.New()
	}
	return s.store.CreateExternalUser(ctx, &ttipb.ExternalUser{
		UserIDs: ids,
		ProviderIDs: ttipb.AuthenticationProviderIdentifiers{
			ProviderID: config.ID,
		},
		ExternalID: claims.Subject,
	})
}

func (s *Server) Login(c echo.Context) error {
	return s.oc.HandleLogin(c)
}

func (s *Server) Callback(c echo.Context) error {
	return s.oc.HandleCallback(c)
}
