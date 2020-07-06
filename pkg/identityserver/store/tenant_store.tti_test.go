// Copyright © 2019 The Things Industries B.V.

package store

import (
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
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
			State:            ttnpb.STATE_APPROVED,
			MaxApplications:  nil,
			MaxClients:       &pbtypes.UInt64Value{Value: 2},
			MaxEndDevices:    &pbtypes.UInt64Value{Value: 3},
			MaxGateways:      &pbtypes.UInt64Value{Value: 4},
			MaxOrganizations: &pbtypes.UInt64Value{Value: 5},
			MaxUsers:         nil,
			Configuration: &ttipb.Configuration{
				DefaultCluster: &ttipb.Configuration_Cluster{
					UI: &ttipb.Configuration_UI{
						BrandingBaseURL: "https://assets.thethings.example.com/branding",
					},
				},
			},
			Billing: &ttipb.Billing{
				Provider: &ttipb.Billing_Stripe_{
					Stripe: &ttipb.Billing_Stripe{
						CustomerID: "cus_XXX",
					},
				},
			},
		})
		a.So(err, should.BeNil)
		a.So(created.TenantID, should.Equal, "foo")
		a.So(created.Name, should.Equal, "Foo Tenant")
		a.So(created.Description, should.Equal, "The Amazing Foo Tenant")
		a.So(created.Attributes, should.HaveLength, 3)
		a.So(created.CreatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
		a.So(created.UpdatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
		a.So(created.State, should.Equal, ttnpb.STATE_APPROVED)
		a.So(created.MaxApplications, should.BeNil)
		a.So(created.MaxClients, should.Resemble, &pbtypes.UInt64Value{Value: 2})
		a.So(created.MaxEndDevices, should.Resemble, &pbtypes.UInt64Value{Value: 3})
		a.So(created.MaxGateways, should.Resemble, &pbtypes.UInt64Value{Value: 4})
		a.So(created.MaxOrganizations, should.Resemble, &pbtypes.UInt64Value{Value: 5})
		a.So(created.MaxUsers, should.BeNil)
		a.So(created.GetConfiguration().GetDefaultCluster().GetUI().GetBrandingBaseURL(), should.Equal, "https://assets.thethings.example.com/branding")
		a.So(created.GetBilling().GetStripe().GetCustomerID(), should.Equal, "cus_XXX")

		got, err := store.GetTenant(ctx, &ttipb.TenantIdentifiers{TenantID: "foo"}, &pbtypes.FieldMask{
			Paths: []string{
				"name", "attributes", "state", "max_applications",
				"max_clients", "max_end_devices", "max_gateways",
				"max_organizations", "max_users",
				"configuration", "billing",
			},
		})
		a.So(err, should.BeNil)
		a.So(got.TenantID, should.Equal, "foo")
		a.So(got.Name, should.Equal, "Foo Tenant")
		a.So(got.Description, should.BeEmpty)
		a.So(got.Attributes, should.HaveLength, 3)
		a.So(got.CreatedAt, should.Equal, created.CreatedAt)
		a.So(got.UpdatedAt, should.Equal, created.UpdatedAt)
		a.So(got.State, should.Equal, ttnpb.STATE_APPROVED)
		a.So(got.MaxApplications, should.BeNil)
		a.So(got.MaxClients, should.Resemble, &pbtypes.UInt64Value{Value: 2})
		a.So(got.MaxEndDevices, should.Resemble, &pbtypes.UInt64Value{Value: 3})
		a.So(got.MaxGateways, should.Resemble, &pbtypes.UInt64Value{Value: 4})
		a.So(got.MaxOrganizations, should.Resemble, &pbtypes.UInt64Value{Value: 5})
		a.So(got.MaxUsers, should.BeNil)
		a.So(got.GetConfiguration().GetDefaultCluster().GetUI().GetBrandingBaseURL(), should.Equal, "https://assets.thethings.example.com/branding")
		a.So(got.GetBilling().GetStripe().GetCustomerID(), should.Equal, "cus_XXX")

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
			State:            ttnpb.STATE_FLAGGED,
			MaxApplications:  &pbtypes.UInt64Value{Value: 2},
			MaxClients:       nil,
			MaxEndDevices:    &pbtypes.UInt64Value{Value: 4},
			MaxGateways:      &pbtypes.UInt64Value{Value: 5},
			MaxOrganizations: nil,
			MaxUsers:         &pbtypes.UInt64Value{Value: 7},
			Billing: &ttipb.Billing{
				Provider: &ttipb.Billing_Stripe_{
					Stripe: &ttipb.Billing_Stripe{
						PlanID: "plan_XXX",
					},
				},
			},
		}, &pbtypes.FieldMask{
			Paths: []string{
				"description", "attributes", "state", "max_applications",
				"max_clients", "max_end_devices", "max_gateways",
				"max_organizations", "max_users",
				"configuration", "billing.provider.stripe.plan_id",
			},
		})
		a.So(err, should.BeNil)
		a.So(updated.Description, should.Equal, "The Amazing Foobar Tenant")
		a.So(updated.Attributes, should.HaveLength, 3)
		a.So(updated.CreatedAt, should.Equal, created.CreatedAt)
		a.So(updated.UpdatedAt, should.HappenAfter, created.CreatedAt)
		a.So(updated.State, should.Equal, ttnpb.STATE_FLAGGED)
		a.So(updated.MaxApplications, should.Resemble, &pbtypes.UInt64Value{Value: 2})
		a.So(updated.MaxClients, should.BeNil)
		a.So(updated.MaxEndDevices, should.Resemble, &pbtypes.UInt64Value{Value: 4})
		a.So(updated.MaxGateways, should.Resemble, &pbtypes.UInt64Value{Value: 5})
		a.So(updated.MaxOrganizations, should.BeNil)
		a.So(updated.MaxUsers, should.Resemble, &pbtypes.UInt64Value{Value: 7})
		a.So(updated.Configuration, should.BeNil)
		a.So(updated.GetBilling().GetStripe().GetPlanID(), should.Equal, "plan_XXX")

		got, err = store.GetTenant(ctx, &ttipb.TenantIdentifiers{TenantID: "foo"}, nil)
		a.So(err, should.BeNil)
		a.So(got.TenantID, should.Equal, created.TenantID)
		a.So(got.Name, should.Equal, created.Name)
		a.So(got.Description, should.Equal, updated.Description)
		a.So(got.Attributes, should.Resemble, updated.Attributes)
		a.So(got.CreatedAt, should.Equal, created.CreatedAt)
		a.So(got.UpdatedAt, should.Equal, updated.UpdatedAt)
		a.So(got.Configuration, should.BeNil)
		a.So(got.GetBilling().GetStripe().GetCustomerID(), should.Equal, "cus_XXX")
		a.So(got.GetBilling().GetStripe().GetPlanID(), should.Equal, "plan_XXX")

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

func TestCountTenantEntities(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()
	tenantID := tenant.FromContext(ctx)

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

		total, err := GetTenantStore(db).CountEntities(ctx, &tenantID, "end_device")
		a.So(err, should.BeNil)
		a.So(total, should.Equal, 1)
	})
}
