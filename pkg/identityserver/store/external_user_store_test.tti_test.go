// Copyright © 2019 The Things Industries B.V.

package store

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestExternalUserStore(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db, &User{}, &Account{}, &ExternalUser{})
		userStore := GetUserStore(db)

		user, err := userStore.CreateUser(ctx, &ttnpb.User{
			UserIdentifiers: ttnpb.UserIdentifiers{UserID: "foo"},
			Name:            "Foo User",
			Description:     "The Amazing Foo User",
		})
		a.So(user, should.NotBeNil)
		a.So(err, should.BeNil)

		store := GetExternalUserStore(db)
		created, err := store.CreateExternalUser(ctx, &ttipb.ExternalUser{
			UserIDs:    user.UserIdentifiers,
			Provider:   ttipb.ExternalUser_OIDC,
			ExternalID: "foo@bar.com",
		})
		a.So(err, should.BeNil)
		if a.So(created, should.NotBeNil) {
			a.So(created.UserIDs.UserID, should.Equal, "foo")
			a.So(created.Provider, should.Equal, ttipb.ExternalUser_OIDC)
			a.So(created.ExternalID, should.Equal, "foo@bar.com")
			a.So(created.CreatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
			a.So(created.UpdatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
		}

		got, err := store.GetExternalUserByUserID(ctx, &ttnpb.UserIdentifiers{UserID: "foo"})
		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.UserIDs.UserID, should.Equal, "foo")
			a.So(got.Provider, should.Equal, ttipb.ExternalUser_OIDC)
			a.So(got.ExternalID, should.Equal, "foo@bar.com")
			a.So(got.CreatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
			a.So(got.UpdatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
		}

		got, err = store.GetExternalUserByExternalID(ctx, "foo@bar.com")
		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.UserIDs.UserID, should.Equal, "foo")
		}

		err = store.DeleteExternalUser(ctx, "foo@bar.com")
		a.So(err, should.BeNil)

		got, err = store.GetExternalUserByExternalID(ctx, "foo@bar.com")
		a.So(err, should.NotBeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)
	})
}
