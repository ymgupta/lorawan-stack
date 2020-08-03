// Copyright © 2020 The Things Industries B.V.

package oidc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/coreos/go-oidc"
	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"golang.org/x/oauth2"
)

type server struct {
	UpstreamServer
	store Store
}

// New creates a new OIDC federated authentication provider wraping the provided upstream.
func New(upstream UpstreamServer, store Store) Server {
	return &server{
		UpstreamServer: upstream,
		store:          store,
	}
}

func (s *server) oidc(ctx context.Context) (*oauth2.Config, *oidc.Provider, error) {
	config := s.GetOIDCConfig(ctx)
	provider, err := oidc.NewProvider(ctx, config.ProviderURL)
	if err != nil {
		return nil, nil, err
	}
	return &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  config.RedirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}, provider, nil
}

const (
	nextKey     = "n"
	oidcNextKey = "_oidc_next"
)

func (s *server) setNextCookie(c echo.Context) {
	ctx := c.Request().Context()
	config := s.GetTemplateData(ctx)
	next := c.QueryParam(nextKey)
	if next == "" {
		next = config.MountPath()
	}
	c.SetCookie(&http.Cookie{
		Name:    oidcNextKey,
		Value:   next,
		Expires: time.Now().Add(24 * time.Hour),
	})
}

func (s *server) goToNext(c echo.Context) error {
	ctx := c.Request().Context()
	config := s.GetTemplateData(ctx)
	next := c.QueryParam(nextKey)
	if next == "" {
		cookie, err := c.Cookie(oidcNextKey)
		if err == nil {
			next = cookie.Value
		} else {
			next = config.MountPath()
		}
	}
	c.SetCookie(&http.Cookie{
		Name:    oidcNextKey,
		Expires: time.Now(),
	})
	return c.Redirect(http.StatusFound, next)
}

// TODO: use pkg/oauthclient
func (s *server) Login(c echo.Context) error {
	ctx := c.Request().Context()
	config, _, err := s.oidc(ctx)
	if err != nil {
		return err
	}
	s.setNextCookie(c)
	return c.Redirect(http.StatusFound, config.AuthCodeURL("banana"))
}

var (
	errInvalidState = errors.DefineInvalidArgument("invalid_state", "invalid state `{state}` provided")
	errNoIDToken    = errors.DefineNotFound("no_id_token", "no ID token found")
)

// TODO: use pkg/oauthclient
func (s *server) Callback(c echo.Context) error {
	ctx := c.Request().Context()
	if state := c.QueryParam("state"); state != "banana" {
		return errInvalidState.WithAttributes("state", state)
	}
	config, provider, err := s.oidc(ctx)
	if err != nil {
		return err
	}
	token, err := config.Exchange(ctx, c.QueryParam("code"))
	if err != nil {
		return err
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return errNoIDToken
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
	return s.goToNext(c)
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
