// Copyright © 2019 The Things Industries B.V.

package stripe_test

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/client"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"google.golang.org/grpc"
)

type mockTenantClientData struct {
	ctx struct {
		Create context.Context
		Get    context.Context
		Update context.Context
	}
	req struct {
		Create *ttipb.CreateTenantRequest
		Get    *ttipb.GetTenantRequest
		Update *ttipb.UpdateTenantRequest
	}
	opts struct {
		Create []grpc.CallOption
		Get    []grpc.CallOption
		Update []grpc.CallOption
	}
	res struct {
		Create *ttipb.Tenant
		Get    *ttipb.Tenant
		Update *ttipb.Tenant
	}
	err struct {
		Create error
		Get    error
		Update error
	}
	tenants map[string]*ttipb.Tenant
}

type mockTenantClient struct {
	mockTenantClientData
	ttipb.TenantRegistryClient
}

var errTenantAlreadyExists = errors.DefineAlreadyExists("tenant_already_exists", "tenant already exists")

func (m *mockTenantClient) Create(ctx context.Context, in *ttipb.CreateTenantRequest, opts ...grpc.CallOption) (*ttipb.Tenant, error) {
	m.ctx.Create, m.req.Create, m.opts.Create = ctx, in, opts
	if m.err.Create != nil || m.res.Create != nil {
		return m.res.Create, m.err.Create
	}
	if _, ok := m.tenants[in.TenantID]; ok {
		return nil, errTenantAlreadyExists.New()
	}
	m.tenants[in.TenantID] = &in.Tenant
	return &in.Tenant, nil
}

var errTenantNotFound = errors.DefineNotFound("tenant_not_found", "tenant not found")

func (m *mockTenantClient) Get(ctx context.Context, in *ttipb.GetTenantRequest, opts ...grpc.CallOption) (*ttipb.Tenant, error) {
	m.ctx.Get, m.req.Get, m.opts.Get = ctx, in, opts
	if m.err.Get != nil || m.res.Get != nil {
		return m.res.Get, m.err.Get
	}
	if tnt, ok := m.tenants[in.TenantID]; ok {
		return tnt, nil
	}
	return nil, errTenantNotFound.New()
}

func (m *mockTenantClient) Update(ctx context.Context, in *ttipb.UpdateTenantRequest, opts ...grpc.CallOption) (*ttipb.Tenant, error) {
	m.ctx.Update, m.req.Update, m.opts.Update = ctx, in, opts
	if m.err.Update != nil || m.res.Update != nil {
		return m.res.Update, m.err.Update
	}
	if _, ok := m.tenants[in.TenantID]; !ok {
		return nil, errTenantNotFound.New()
	}
	m.tenants[in.TenantID] = &in.Tenant
	return &in.Tenant, nil
}

type mockStripeBackend struct {
	stripe.Backend

	CallMock func(method, path, key string, params stripe.ParamsContainer, v interface{}) error
}

func (m *mockStripeBackend) Call(method, path, key string, params stripe.ParamsContainer, v interface{}) error {
	if m.CallMock != nil {
		return m.CallMock(method, path, key, params, v)
	}
	panic("CallMock is nil")
}

func createStripeMock(ctx context.Context, apiKey string, mock *mockStripeBackend) *client.API {
	backends := stripe.NewBackends(nil)
	backends.API = mock
	return client.New(apiKey, backends)
}

func createHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}
