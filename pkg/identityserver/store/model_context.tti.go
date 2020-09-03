// Copyright © 2019 The Things Industries B.V.

//+build tti

package store

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
)

// SetContext needs to be called before creating models.
func (m *Model) SetContext(ctx context.Context) {
	m.ctx = ctx
}

// SetContext needs to be called before creating models.
func (a *Account) SetContext(ctx context.Context) {
	a.TenantID = tenant.FromContext(ctx).TenantID
	a.Model.SetContext(ctx)
}

// SetContext needs to be called before creating models.
func (k *APIKey) SetContext(ctx context.Context) {
	k.TenantID = tenant.FromContext(ctx).TenantID
	k.Model.SetContext(ctx)
}

// SetContext needs to be called before creating models.
func (app *Application) SetContext(ctx context.Context) {
	app.TenantID = tenant.FromContext(ctx).TenantID
	app.Model.SetContext(ctx)
}

// SetContext needs to be called before creating models.
func (cli *Client) SetContext(ctx context.Context) {
	if tenantID := tenant.FromContext(ctx).TenantID; tenantID != "" {
		cli.TenantID = &tenantID
	}
	cli.Model.SetContext(ctx)
}

// SetContext needs to be called before creating models.
func (dev *EndDevice) SetContext(ctx context.Context) {
	dev.TenantID = tenant.FromContext(ctx).TenantID
	dev.Model.SetContext(ctx)
}

// SetContext needs to be called before creating models.
func (gtw *Gateway) SetContext(ctx context.Context) {
	gtw.TenantID = tenant.FromContext(ctx).TenantID
	gtw.Model.SetContext(ctx)
}

// SetContext needs to be called before creating models.
func (i *Invitation) SetContext(ctx context.Context) {
	i.TenantID = tenant.FromContext(ctx).TenantID
	i.Model.SetContext(ctx)
}

// SetContext needs to be called before creating models.
func (c *AuthorizationCode) SetContext(ctx context.Context) {
	c.TenantID = tenant.FromContext(ctx).TenantID
	c.Model.SetContext(ctx)
}

// SetContext needs to be called before creating models.
func (t *AccessToken) SetContext(ctx context.Context) {
	t.TenantID = tenant.FromContext(ctx).TenantID
	t.Model.SetContext(ctx)
}

// SetContext sets the context on the organization model and the embedded account model.
func (org *Organization) SetContext(ctx context.Context) {
	org.Model.SetContext(ctx)
	org.Account.SetContext(ctx)
}

// SetContext sets the context on both the Model and Account.
func (usr *User) SetContext(ctx context.Context) {
	usr.TenantID = tenant.FromContext(ctx).TenantID
	usr.Model.SetContext(ctx)
	usr.Account.SetContext(ctx)
}

// SetContext sets the context on the external user model.
func (eu *ExternalUser) SetContext(ctx context.Context) {
	eu.TenantID = tenant.FromContext(ctx).TenantID
	eu.Model.SetContext(ctx)
}

// SetContext sets the context on the authentication provider model.
func (ap *AuthenticationProvider) SetContext(ctx context.Context) {
	ap.TenantID = tenant.FromContext(ctx).TenantID
	ap.Model.SetContext(ctx)
}
