// Copyright © 2020 The Things Industries B.V.

package oidc

import (
	"context"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/webui"
)

// Store used by the OIDC federated authentication provider.
type Store interface {
	store.UserStore
	store.ExternalUserStore
}

// UpstreamServer is an upstream authentication server.
type UpstreamServer interface {
	GetTemplateData(context.Context) webui.TemplateData
	CreateUserSession(echo.Context, ttnpb.UserIdentifiers) error
}
