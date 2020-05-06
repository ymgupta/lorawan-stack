// Copyright © 2019 The Things Industries B.V.

package license_test

import (
	"context"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	. "go.thethings.network/lorawan-stack/v3/pkg/license"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

func TestContext(t *testing.T) {
	t.Run("Background Context", func(t *testing.T) {
		a := assertions.New(t)
		ctx := context.Background()

		l := FromContext(ctx)

		a.So(l.LicenseID, should.Equal, "testing")

		a.So(RequireComponent(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY), should.BeNil)
		a.So(RequireMultiTenancy(ctx), should.BeNil)
		a.So(RequireComponentAddress(ctx, "localhost"), should.BeNil)
		a.So(RequireComponentAddress(ctx, "https://localhost:8885/console"), should.BeNil)
		a.So(RequireDevAddrPrefix(ctx, types.DevAddrPrefix{}), should.BeNil)
		a.So(RequireJoinEUIPrefix(ctx, types.EUI64Prefix{}), should.BeNil)
	})

	t.Run("Expired License", func(t *testing.T) {
		a := assertions.New(t)
		testLicense := FromContext(test.Context())
		testLicense.ValidUntil = time.Now()
		ctx := NewContextWithLicense(test.Context(), testLicense)

		a.So(RequireComponent(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY), should.NotBeNil)
		a.So(RequireMultiTenancy(ctx), should.NotBeNil)
		a.So(RequireComponentAddress(ctx, "localhost"), should.NotBeNil)
		a.So(RequireDevAddrPrefix(ctx, types.DevAddrPrefix{}), should.NotBeNil)
		a.So(RequireJoinEUIPrefix(ctx, types.EUI64Prefix{}), should.NotBeNil)
	})

	t.Run("Restricted Components", func(t *testing.T) {
		a := assertions.New(t)
		testLicense := FromContext(test.Context())
		testLicense.Components = []ttnpb.ClusterRole{
			ttnpb.ClusterRole_ENTITY_REGISTRY,
		}
		ctx := NewContextWithLicense(test.Context(), testLicense)

		a.So(RequireComponent(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY), should.BeNil)
		a.So(RequireComponent(ctx, ttnpb.ClusterRole_CRYPTO_SERVER), should.NotBeNil)
	})

	t.Run("Multi Tenancy", func(t *testing.T) {
		a := assertions.New(t)
		testLicense := FromContext(test.Context())
		testLicense.MultiTenancy = true
		ctx := NewContextWithLicense(test.Context(), testLicense)

		a.So(RequireMultiTenancy(ctx), should.BeNil)
	})

	t.Run("No Multi Tenancy", func(t *testing.T) {
		a := assertions.New(t)
		testLicense := FromContext(test.Context())
		testLicense.MultiTenancy = false
		ctx := NewContextWithLicense(test.Context(), testLicense)

		a.So(RequireMultiTenancy(ctx), should.NotBeNil)
	})

	t.Run("Restricted Addresses", func(t *testing.T) {
		a := assertions.New(t)
		testLicense := FromContext(test.Context())
		testLicense.ComponentAddressRegexps = []string{
			`^([a-z0-9](?:[-]?[a-z0-9]){2,}\.)?localhost$`,
		}
		ctx := NewContextWithLicense(test.Context(), testLicense)

		a.So(RequireComponentAddress(ctx, "localhost"), should.BeNil)
		a.So(RequireComponentAddress(ctx, "https://localhost:8885/console"), should.BeNil)
		a.So(RequireComponentAddress(ctx, "foo.localhost"), should.BeNil)
		a.So(RequireComponentAddress(ctx, "https://foo.localhost:8885/console"), should.BeNil)
		a.So(RequireComponentAddress(ctx, "localhost.com"), should.NotBeNil)
		a.So(RequireComponentAddress(ctx, "other.foo.localhost"), should.NotBeNil)
	})

	t.Run("Restricted DevAddr Prefixes", func(t *testing.T) {
		a := assertions.New(t)
		testLicense := FromContext(test.Context())
		testLicense.DevAddrPrefixes = []types.DevAddrPrefix{
			{DevAddr: types.DevAddr{0x12, 0x34, 0x56, 0x70}, Length: 28},
		}
		ctx := NewContextWithLicense(test.Context(), testLicense)

		a.So(RequireDevAddrPrefix(ctx, types.DevAddrPrefix{
			DevAddr: types.DevAddr{0x12, 0x34, 0x56, 0x70}, Length: 28},
		), should.BeNil)
		a.So(RequireDevAddrPrefix(ctx, types.DevAddrPrefix{
			DevAddr: types.DevAddr{0x12, 0x34, 0x56, 0x78}, Length: 29},
		), should.BeNil)
		a.So(RequireDevAddrPrefix(ctx, types.DevAddrPrefix{
			DevAddr: types.DevAddr{0x12, 0x34, 0x56, 0x80}, Length: 28},
		), should.NotBeNil)
	})

	t.Run("Restricted JoinEUI Prefixes", func(t *testing.T) {
		a := assertions.New(t)
		testLicense := FromContext(test.Context())
		testLicense.JoinEUIPrefixes = []types.EUI64Prefix{
			{EUI64: types.EUI64{0x12, 0x34, 0x56, 0x70, 0x00, 0x00, 0x00, 0x00}, Length: 28},
		}
		ctx := NewContextWithLicense(test.Context(), testLicense)

		a.So(RequireJoinEUIPrefix(ctx, types.EUI64Prefix{
			EUI64: types.EUI64{0x12, 0x34, 0x56, 0x70, 0x00, 0x00, 0x00, 0x00}, Length: 28},
		), should.BeNil)
		a.So(RequireJoinEUIPrefix(ctx, types.EUI64Prefix{
			EUI64: types.EUI64{0x12, 0x34, 0x56, 0x78, 0x00, 0x00, 0x00, 0x00}, Length: 29},
		), should.BeNil)
		a.So(RequireJoinEUIPrefix(ctx, types.EUI64Prefix{
			EUI64: types.EUI64{0x12, 0x34, 0x56, 0x80, 0x00, 0x00, 0x00, 0x00}, Length: 28},
		), should.NotBeNil)
	})
}
