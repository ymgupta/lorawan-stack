// Copyright © 2020 The Things Industries B.V.

package store

import (
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestAuthenticationProviderStore(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db, &AuthenticationProvider{})

		store := GetAuthenticationProviderStore(db)
		created, err := store.CreateAuthenticationProvider(ctx, &ttipb.AuthenticationProvider{
			AuthenticationProviderIdentifiers: ttipb.AuthenticationProviderIdentifiers{
				ProviderID: "foo",
			},
			Name:               "bar",
			AllowRegistrations: true,
			Configuration: &ttipb.AuthenticationProvider_Configuration{
				Provider: &ttipb.AuthenticationProvider_Configuration_OIDC{
					OIDC: &ttipb.AuthenticationProvider_OIDC{
						ClientID:     "foo-client",
						ClientSecret: "foo-secret",
						ProviderURL:  "https://foo.bar",
					},
				},
			},
		})
		a.So(err, should.BeNil)
		if a.So(created, should.NotBeNil) {
			a.So(created.ProviderID, should.Equal, "foo")
			a.So(created.Name, should.Equal, "bar")
			a.So(created.AllowRegistrations, should.BeTrue)
			if oidc := created.Configuration.GetOIDC(); a.So(oidc, should.NotBeNil) {
				a.So(oidc.ClientID, should.Equal, "foo-client")
				a.So(oidc.ClientSecret, should.Equal, "foo-secret")
				a.So(oidc.ProviderURL, should.Equal, "https://foo.bar")
			}
			a.So(created.CreatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
			a.So(created.UpdatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
		}

		got, err := store.GetAuthenticationProvider(ctx,
			&ttipb.AuthenticationProviderIdentifiers{ProviderID: "foo"},
			&types.FieldMask{Paths: []string{"name", "configuration.provider.oidc.client_id"}})
		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.ProviderID, should.Equal, "foo")
			a.So(got.Name, should.Equal, "bar")
			a.So(got.AllowRegistrations, should.BeFalse)
			if oidc := got.Configuration.GetOIDC(); a.So(oidc, should.NotBeNil) {
				a.So(oidc.ClientID, should.Equal, "foo-client")
				a.So(oidc.ClientSecret, should.Equal, "")
				a.So(oidc.ProviderURL, should.Equal, "")
			}
			a.So(got.CreatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
			a.So(got.UpdatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
		}

		list, err := store.FindAuthenticationProviders(ctx,
			[]*ttipb.AuthenticationProviderIdentifiers{
				{ProviderID: "foo"},
			},
			&types.FieldMask{
				Paths: []string{"name", "allow_registrations", "configuration.provider"},
			})
		a.So(err, should.BeNil)
		if a.So(list, should.HaveLength, 1) {
			a.So(list[0].ProviderID, should.Equal, "foo")
			a.So(list[0].Name, should.Equal, "bar")
			a.So(list[0].AllowRegistrations, should.BeTrue)
			if oidc := list[0].Configuration.GetOIDC(); a.So(oidc, should.NotBeNil) {
				a.So(oidc.ClientID, should.Equal, "foo-client")
				a.So(oidc.ClientSecret, should.Equal, "foo-secret")
				a.So(oidc.ProviderURL, should.Equal, "https://foo.bar")
			}
			a.So(list[0].CreatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
			a.So(list[0].UpdatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
		}

		err = store.DeleteAuthenticationProvider(ctx, &ttipb.AuthenticationProviderIdentifiers{ProviderID: "foo"})
		a.So(err, should.BeNil)

		got, err = store.GetAuthenticationProvider(ctx, &ttipb.AuthenticationProviderIdentifiers{ProviderID: "foo"}, nil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)

		list, err = store.FindAuthenticationProviders(ctx, nil, nil)
		a.So(err, should.BeNil)
		a.So(list, should.BeEmpty)
	})
}
