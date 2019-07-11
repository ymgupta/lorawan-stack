// Copyright © 2019 The Things Industries B.V.

//+build tti

package unique_test

import (
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	. "go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestValidity(t *testing.T) {
	eui := types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}
	for _, tc := range []ttnpb.Identifiers{
		nil,
		(*ttnpb.ApplicationIdentifiers)(nil),
		ttnpb.ApplicationIdentifiers{},
		(*ttnpb.ClientIdentifiers)(nil),
		ttnpb.ClientIdentifiers{},
		(*ttnpb.EndDeviceIdentifiers)(nil),
		ttnpb.EndDeviceIdentifiers{},
		&ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "foo"},
			DevEUI:                 &eui,
		},
		ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "foo"},
			DevEUI:                 &eui,
		},
		(*ttnpb.GatewayIdentifiers)(nil),
		ttnpb.GatewayIdentifiers{},
		&ttnpb.GatewayIdentifiers{
			EUI: &eui,
		},
		ttnpb.GatewayIdentifiers{
			EUI: &eui,
		},
		(*ttnpb.OrganizationIdentifiers)(nil),
		ttnpb.OrganizationIdentifiers{},
		(*ttnpb.UserIdentifiers)(nil),
		ttnpb.UserIdentifiers{},
	} {
		t.Run(fmt.Sprintf("%T", tc), func(t *testing.T) {
			a := assertions.New(t)
			a.So(func() { ID(test.Context(), tc) }, should.Panic)
		})
	}
}

func TestToTenantID(t *testing.T) {
	for _, tc := range []struct {
		Name     string
		UID      string
		TenantID string
	}{
		{
			Name:     "ApplicationWithTenant",
			UID:      "foo-application@foo-tenant",
			TenantID: "foo-tenant",
		},
		{
			Name:     "EndDeviceWithTenant",
			UID:      "foo-application.foo-device@foo-tenant",
			TenantID: "foo-tenant",
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			id, err := ToTenantID(tc.UID)
			a.So(err, should.BeNil)
			a.So(id, should.Resemble, ttipb.TenantIdentifiers{TenantID: tc.TenantID})
		})
	}
}

func TestRoundtrip(t *testing.T) {
	for _, tc := range []struct {
		ID       ttnpb.Identifiers
		Expected string
		Parser   func(string) (ttnpb.Identifiers, error)
	}{
		{
			ttnpb.ApplicationIdentifiers{ApplicationID: "foo"},
			"foo@foo-tenant",
			func(uid string) (ttnpb.Identifiers, error) { return ToApplicationID(uid) },
		},
		{
			&ttnpb.ApplicationIdentifiers{ApplicationID: "foo"},
			"foo@foo-tenant",
			func(uid string) (ttnpb.Identifiers, error) { return ToApplicationID(uid) },
		},
		{
			ttnpb.ClientIdentifiers{ClientID: "foo"},
			"foo@foo-tenant",
			func(uid string) (ttnpb.Identifiers, error) { return ToClientID(uid) },
		},
		{
			&ttnpb.ClientIdentifiers{ClientID: "foo"},
			"foo@foo-tenant",
			func(uid string) (ttnpb.Identifiers, error) { return ToClientID(uid) },
		},
		{
			ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: "foo-app",
				},
				DeviceID: "foo-device",
			},
			"foo-app.foo-device@foo-tenant",
			func(uid string) (ttnpb.Identifiers, error) { return ToDeviceID(uid) },
		},
		{
			&ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: "foo-app",
				},
				DeviceID: "foo-device",
			},
			"foo-app.foo-device@foo-tenant",
			func(uid string) (ttnpb.Identifiers, error) { return ToDeviceID(uid) },
		},
		{
			ttnpb.GatewayIdentifiers{GatewayID: "foo"},
			"foo@foo-tenant",
			func(uid string) (ttnpb.Identifiers, error) { return ToGatewayID(uid) },
		},
		{
			&ttnpb.GatewayIdentifiers{GatewayID: "foo"},
			"foo@foo-tenant",
			func(uid string) (ttnpb.Identifiers, error) { return ToGatewayID(uid) },
		},
		{
			ttnpb.OrganizationIdentifiers{OrganizationID: "foo"},
			"foo@foo-tenant",
			func(uid string) (ttnpb.Identifiers, error) { return ToOrganizationID(uid) },
		},
		{
			&ttnpb.OrganizationIdentifiers{OrganizationID: "foo"},
			"foo@foo-tenant",
			func(uid string) (ttnpb.Identifiers, error) { return ToOrganizationID(uid) },
		},
		{
			ttnpb.UserIdentifiers{UserID: "foo"},
			"foo@foo-tenant",
			func(uid string) (ttnpb.Identifiers, error) { return ToUserID(uid) },
		},
		{
			&ttnpb.UserIdentifiers{UserID: "foo"},
			"foo@foo-tenant",
			func(uid string) (ttnpb.Identifiers, error) { return ToUserID(uid) },
		},
	} {
		t.Run(fmt.Sprintf("%T", tc.ID), func(t *testing.T) {
			a := assertions.New(t)
			a.So(ID(test.Context(), tc.ID), should.Equal, tc.Expected)
			if id, ok := tc.ID.(interface {
				EntityIdentifiers() *ttnpb.EntityIdentifiers
			}); ok {
				wrapped := id.EntityIdentifiers()
				a.So(ID(test.Context(), wrapped), should.Equal, tc.Expected)
				a.So(ID(test.Context(), *wrapped), should.Equal, tc.Expected)
			}
			a.So(ID(test.Context(), tc.ID), should.Equal, tc.Expected)
			if tc.Parser != nil {
				parsed, err := tc.Parser(tc.Expected)
				if a.So(err, should.BeNil) {
					a.So(parsed.CombinedIdentifiers(), should.Resemble, tc.ID.CombinedIdentifiers())
				}
			}
		})
	}
}

func TestValidatorForIdentifiers(t *testing.T) {
	a := assertions.New(t)
	for _, run := range []struct {
		Name   string
		ID     func(string) ttnpb.Identifiers
		Parser func(string) (ttnpb.Identifiers, error)
	}{
		{
			"ApplicationID",
			func(uid string) ttnpb.Identifiers { return ttnpb.ApplicationIdentifiers{ApplicationID: uid} },
			func(uid string) (ttnpb.Identifiers, error) { return ToApplicationID(uid) },
		},
		{
			"ClientID",
			func(uid string) ttnpb.Identifiers { return ttnpb.ClientIdentifiers{ClientID: uid} },
			func(uid string) (ttnpb.Identifiers, error) { return ToClientID(uid) },
		},
		{
			"GatewayID",
			func(uid string) ttnpb.Identifiers { return ttnpb.GatewayIdentifiers{GatewayID: uid} },
			func(uid string) (ttnpb.Identifiers, error) { return ToGatewayID(uid) },
		},
		{
			"OrganizationID",
			func(uid string) ttnpb.Identifiers { return ttnpb.OrganizationIdentifiers{OrganizationID: uid} },
			func(uid string) (ttnpb.Identifiers, error) { return ToOrganizationID(uid) },
		},
		{
			"UserID",
			func(uid string) ttnpb.Identifiers { return ttnpb.UserIdentifiers{UserID: uid} },
			func(uid string) (ttnpb.Identifiers, error) { return ToUserID(uid) },
		},
	} {
		for _, tc := range []struct {
			Name          string
			InputUID      string
			ExpectedError func(error) bool
		}{
			{
				"ValidID",
				"test",
				nil,
			},
			{
				"ValidMinLength",
				"oza",
				nil,
			},
			{
				"ValidMaxLength",
				"ozaj8qs0sait7oudxqbfyx6b14yuahcfrdlb",
				nil,
			},
			{
				"ValidNumerics",
				"1id1",
				nil,
			},
			{
				"InvalidEmptyString",
				"",
				errors.IsInvalidArgument,
			},
			{
				"InvalidDashes",
				"-id",
				errors.IsInvalidArgument,
			},
			{
				"InvalidDashes1",
				"id-",
				errors.IsInvalidArgument,
			},
			{
				"InvalidUnderscore",
				"id_test",
				errors.IsInvalidArgument,
			},
			{
				"InvalidDollar",
				"id$test",
				errors.IsInvalidArgument,
			},
			{
				"InvalidMinLength",
				"id",
				errors.IsInvalidArgument,
			},
			{
				"InvalidMaxLength",
				"ozaj8qs0sait7oudxqbfyx6b14yuahcfrdlbh",
				errors.IsInvalidArgument,
			},
		} {
			t.Run(fmt.Sprintf("%s/%s", run.Name, tc.Name), func(t *testing.T) {
				if run.Parser == nil {
					t.Fatal("Parser Not Defined")
				}
				parsed, err := run.Parser(tc.InputUID + "@foo-tenant")
				if tc.ExpectedError == nil {
					a.So(err, should.BeNil)
					if !a.So(parsed, should.Resemble, run.ID(tc.InputUID)) {
						t.FailNow()
					}
				} else {
					if !a.So(tc.ExpectedError(err), should.BeTrue) {
						t.FailNow()
					}
				}
			})
		}
	}

}

func TestValidatorForDeviceIDs(t *testing.T) {
	a := assertions.New(t)
	for _, tc := range []struct {
		Name               string
		InputUID           string
		ExpectedIdentifier ttnpb.EndDeviceIdentifiers
		ExpectedError      func(error) bool
	}{
		{
			"ValidID",
			"foo-app.foo-device",
			ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: "foo-app",
				},
				DeviceID: "foo-device",
			},
			nil,
		},
		{
			"ValidAppIDValidMinLength",
			"foo-app.foo",
			ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: "foo-app",
				},
				DeviceID: "foo",
			},
			nil,
		},
		{
			"ValidAppIDValidMaxLength",
			"foo-app.ozaj8qs0sait7oudxqbfyx6b14yuahcfrdlb",
			ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: "foo-app",
				},
				DeviceID: "ozaj8qs0sait7oudxqbfyx6b14yuahcfrdlb",
			},
			nil,
		},
		{
			"ValidAppIDValidNumerics",
			"foo-app.1d1",
			ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: "foo-app",
				},
				DeviceID: "1d1",
			},
			nil,
		},
		{
			"InvalidFormat",
			"foo-appfoo-device",
			ttnpb.EndDeviceIdentifiers{},
			errors.IsInvalidArgument,
		},
		{
			"InvalidAppID",
			"foo_app.foo-device",
			ttnpb.EndDeviceIdentifiers{},
			errors.IsInvalidArgument,
		},
		{
			"ValidAppIDInvalidEmptyString",
			"foo-app.",
			ttnpb.EndDeviceIdentifiers{},
			errors.IsInvalidArgument,
		},
		{
			"ValidAppIDInvalidEmptyString",
			"foo-app.",
			ttnpb.EndDeviceIdentifiers{},
			errors.IsInvalidArgument,
		},
		{
			"ValidAppIDInvalidDashes",
			"foo-app.-foo",
			ttnpb.EndDeviceIdentifiers{},
			errors.IsInvalidArgument,
		},
		{
			"ValidAppIDInvalidDashes1",
			"foo-app.foo-",
			ttnpb.EndDeviceIdentifiers{},
			errors.IsInvalidArgument,
		},
		{
			"ValidAppIDInvalidUnderscore",
			"foo-app.foo_device",
			ttnpb.EndDeviceIdentifiers{},
			errors.IsInvalidArgument,
		},
		{
			"ValidAppIDInvalidDot",
			"foo-app.foo.device",
			ttnpb.EndDeviceIdentifiers{},
			errors.IsInvalidArgument,
		},
		{
			"ValidAppIDInvalidMinLength",
			"foo-app.id",
			ttnpb.EndDeviceIdentifiers{},
			errors.IsInvalidArgument,
		},
		{
			"ValidAppIDInvalidMaxLength",
			"foo-app.ozaj8qs0sait7oudxqbfyx6b14yuahcfrdlbh",
			ttnpb.EndDeviceIdentifiers{},
			errors.IsInvalidArgument,
		},
	} {
		t.Run(fmt.Sprintf("%s", tc.Name), func(t *testing.T) {
			devIDs, err := ToDeviceID(tc.InputUID + "@foo-tenant")
			if tc.ExpectedError == nil {
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				if !a.So(devIDs, should.Resemble, tc.ExpectedIdentifier) {
					t.FailNow()
				}
			} else {
				if !a.So(tc.ExpectedError(err), should.BeTrue) {
					t.FailNow()
				}
			}
		})
	}
}
