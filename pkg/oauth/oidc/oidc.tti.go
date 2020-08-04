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

type server struct {
	UpstreamServer
	store Store
	oc    *oauthclient.OAuthClient
}

// New creates a new OIDC federated authentication provider wraping the provided upstream.
func New(c *component.Component, upstream UpstreamServer, store Store) (Server, error) {
	s := &server{
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

func (s *server) oidc(ctx context.Context) (*oauth2.Config, *oidc.Provider, error) {
	config := s.GetOIDCConfig(ctx)
	provider, err := oidc.NewProvider(ctx, config.ProviderURL)
	if err != nil {
		return nil, nil, err
	}
	uiConfig := s.GetTemplateData(ctx)
	redirectURL := fmt.Sprintf("%s/login/oidc/callback", strings.TrimSuffix(uiConfig.CanonicalURL, "/"))
	return &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}, provider, nil
}

func (s *server) oauth(c echo.Context) *oauth2.Config {
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

func (s *server) callback(c echo.Context, token *oauth2.Token, next string) error {
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
	externalUser, err := s.store.GetExternalUserByExternalID(ctx, claims.Subject)
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

func (s *server) username(ct *claimsToken) string {
	return fmt.Sprintf("oidc%v", ct.Subject)
}

var errRegistrationsDisabled = errors.DefineFailedPrecondition("registrations_disabled", "account registrations are disabled")

func (s *server) createExternalUser(ctx context.Context, claims *claimsToken) (*ttipb.ExternalUser, error) {
	config := s.GetOIDCConfig(ctx)
	if !config.AllowRegistrations {
		return nil, errRegistrationsDisabled.New()
	}
	ids := ttnpb.UserIdentifiers{UserID: s.username(claims)}
	now := time.Now()
	password, err := auth.Hash(ctx, random.String(64))
	if err != nil {
		return nil, err
	}
	newUser := &ttnpb.User{
		UserIdentifiers:                ids,
		Name:                           claims.Name,
		PrimaryEmailAddress:            claims.Email,
		PrimaryEmailAddressValidatedAt: &now,
		Password:                       password,
		PasswordUpdatedAt:              &now,
		State:                          ttnpb.STATE_APPROVED,
	}
	if _, err := s.store.CreateUser(ctx, newUser); err != nil {
		return nil, err
	}
	return s.store.CreateExternalUser(ctx, &ttipb.ExternalUser{
		UserIDs:    ids,
		Provider:   ttipb.ExternalUser_OIDC,
		ExternalID: claims.Subject,
	})
}

func (s *server) Login(c echo.Context) error {
	return s.oc.HandleLogin(c)
}

func (s *server) Callback(c echo.Context) error {
	return s.oc.HandleCallback(c)
}
