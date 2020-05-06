// Copyright © 2019 The Things Industries B.V.

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
)

func TestHTTPMiddleware(t *testing.T) {
	m := Middleware(tenant.Config{
		DefaultID: "default",
		BaseDomains: []string{
			"nz.cluster.ttn",
			"identity.ttn",
		},
	})

	testCases := []struct {
		desc string
		req  func() *http.Request
	}{
		{
			desc: "Host",
			req: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "https://foo-bar.identity.ttn/", nil)
				return req
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			a := assertions.New(t)

			rec := httptest.NewRecorder()
			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				a.So(tenant.FromContext(r.Context()).TenantID, should.Equal, "foo-bar")
			})).ServeHTTP(rec, tc.req())

			res := rec.Result()

			a.So(res.StatusCode, should.Equal, http.StatusOK)
		})
	}

	redirectTestCases := []struct {
		desc     string
		req      func() *http.Request
		redirect string
	}{
		{
			desc: "Host",
			req: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "http://identity.ttn/path/to?query=true", nil)
				return req
			},
			redirect: "http://default.identity.ttn/path/to?query=true",
		},
		{
			desc: "HTTPSHost",
			req: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "https://identity.ttn/path/to?query=true", nil)
				return req
			},
			redirect: "https://default.identity.ttn/path/to?query=true",
		},
		{
			desc: "HostPort",
			req: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "http://identity.ttn:1885/path/to?query=true", nil)
				return req
			},
			redirect: "http://default.identity.ttn:1885/path/to?query=true",
		},
		{
			desc: "HTTPSHostPort",
			req: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "https://identity.ttn:8885/path/to?query=true", nil)
				return req
			},
			redirect: "https://default.identity.ttn:8885/path/to?query=true",
		},
	}
	for _, tc := range redirectTestCases {
		t.Run(tc.desc, func(t *testing.T) {
			a := assertions.New(t)

			rec := httptest.NewRecorder()

			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("Wrapped handler was called when it should not have been")
			})).ServeHTTP(rec, tc.req())

			res := rec.Result()

			a.So(res.StatusCode, should.Equal, http.StatusFound)
			a.So(rec.Header().Get("Location"), should.Equal, tc.redirect)
		})
	}
}
