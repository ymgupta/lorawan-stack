// Copyright © 2019 The Things Industries B.V.

package store

import (
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

func TestTenantStore(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db, &Tenant{}, &Attribute{})
		store := GetTenantStore(db)

		created, err := store.CreateTenant(ctx, &ttipb.Tenant{
			TenantIdentifiers: ttipb.TenantIdentifiers{TenantID: "foo"},
			Name:              "Foo Tenant",
			Description:       "The Amazing Foo Tenant",
			Attributes: map[string]string{
				"foo": "bar",
				"bar": "baz",
				"baz": "qux",
			},
		})
		a.So(err, should.BeNil)
		a.So(created.TenantID, should.Equal, "foo")
		a.So(created.Name, should.Equal, "Foo Tenant")
		a.So(created.Description, should.Equal, "The Amazing Foo Tenant")
		a.So(created.Attributes, should.HaveLength, 3)
		a.So(created.CreatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
		a.So(created.UpdatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))

		got, err := store.GetTenant(ctx, &ttipb.TenantIdentifiers{TenantID: "foo"}, &pbtypes.FieldMask{Paths: []string{"name", "attributes"}})
		a.So(err, should.BeNil)
		a.So(got.TenantID, should.Equal, "foo")
		a.So(got.Name, should.Equal, "Foo Tenant")
		a.So(got.Description, should.BeEmpty)
		a.So(got.Attributes, should.HaveLength, 3)
		a.So(got.CreatedAt, should.Equal, created.CreatedAt)
		a.So(got.UpdatedAt, should.Equal, created.UpdatedAt)

		_, err = store.UpdateTenant(ctx, &ttipb.Tenant{
			TenantIdentifiers: ttipb.TenantIdentifiers{TenantID: "bar"},
		}, nil)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		updated, err := store.UpdateTenant(ctx, &ttipb.Tenant{
			TenantIdentifiers: ttipb.TenantIdentifiers{TenantID: "foo"},
			Name:              "Foobar Tenant",
			Description:       "The Amazing Foobar Tenant",
			Attributes: map[string]string{
				"foo": "bar",
				"baz": "baz",
				"qux": "foo",
			},
		}, &pbtypes.FieldMask{Paths: []string{"description", "attributes"}})
		a.So(err, should.BeNil)
		a.So(updated.Description, should.Equal, "The Amazing Foobar Tenant")
		a.So(updated.Attributes, should.HaveLength, 3)
		a.So(updated.CreatedAt, should.Equal, created.CreatedAt)
		a.So(updated.UpdatedAt, should.HappenAfter, created.CreatedAt)

		got, err = store.GetTenant(ctx, &ttipb.TenantIdentifiers{TenantID: "foo"}, nil)
		a.So(err, should.BeNil)
		a.So(got.TenantID, should.Equal, created.TenantID)
		a.So(got.Name, should.Equal, created.Name)
		a.So(got.Description, should.Equal, updated.Description)
		a.So(got.Attributes, should.Resemble, updated.Attributes)
		a.So(got.CreatedAt, should.Equal, created.CreatedAt)
		a.So(got.UpdatedAt, should.Equal, updated.UpdatedAt)

		list, err := store.FindTenants(ctx, nil, &pbtypes.FieldMask{Paths: []string{"name"}})
		a.So(err, should.BeNil)
		if a.So(list, should.HaveLength, 1) {
			a.So(list[0].Name, should.EndWith, got.Name)
		}

		err = store.DeleteTenant(ctx, &ttipb.TenantIdentifiers{TenantID: "foo"})
		a.So(err, should.BeNil)

		got, err = store.GetTenant(ctx, &ttipb.TenantIdentifiers{TenantID: "foo"}, nil)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		list, err = store.FindTenants(ctx, nil, nil)
		a.So(err, should.BeNil)
		a.So(list, should.BeEmpty)
	})
}

func TestGetTenantIDForGatewayEUI(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db, &Tenant{}, &Gateway{})

		eui := types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}

		_, err := GetGatewayStore(db).CreateGateway(ctx, &ttnpb.Gateway{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{
				GatewayID: "foo",
				EUI:       &eui,
			},
			Name:        "Foo Gateway",
			Description: "The Amazing Foo Gateway",
		})
		a.So(err, should.BeNil)

		id, err := GetTenantStore(db).GetTenantIDForGatewayEUI(ctx, eui)
		a.So(err, should.BeNil)

		a.So(*id, should.Resemble, tenant.FromContext(ctx))
	})
}

func TestGetTenantIDForEndDeviceEUIs(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db, &Tenant{}, &EndDevice{})

		joinEUI := types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}
		devEUI := types.EUI64{8, 7, 6, 5, 4, 3, 2, 1}

		_, err := GetEndDeviceStore(db).CreateEndDevice(ctx, &ttnpb.EndDevice{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: "foo-app",
				},
				DeviceID: "foo",
				JoinEUI:  &joinEUI,
				DevEUI:   &devEUI,
			},
			Name: "Foo Device",
		})
		a.So(err, should.BeNil)

		id, err := GetTenantStore(db).GetTenantIDForEndDeviceEUIs(ctx, joinEUI, devEUI)
		a.So(err, should.BeNil)

		a.So(*id, should.Resemble, tenant.FromContext(ctx))
	})
}
