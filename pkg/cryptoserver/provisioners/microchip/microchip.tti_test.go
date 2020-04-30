// Copyright © 2019 The Things Industries B.V.

package microchip_test

import (
	"context"
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/cryptoserver/provisioners/microchip"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func eui64Ptr(v types.EUI64) *types.EUI64 { return &v }

func TestMicrochip(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	fn := func(ctx context.Context) (map[string]Key, error) {
		return map[string]Key{
			"test": {0xE0, 0xE1, 0xE2, 0xE3, 0xE4, 0xE5, 0xE6, 0xE7, 0xE8, 0xE9, 0xEA, 0xEB, 0xEC, 0xED, 0xEE, 0xEF, 0xF0, 0xF1, 0xF2, 0xF3, 0xF4, 0xF5, 0xF6, 0xF7, 0xF8, 0xF9, 0xFA, 0xFB, 0xFC, 0xFD, 0xFE, 0xFF},
		}, nil
	}

	conf := &Config{
		LoadKeysFunc: fn,
	}
	svc, err := New(ctx, conf)
	a.So(err, should.BeNil)

	dev := &ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "foo-app",
			},
			DeviceID: "foo-device",
			JoinEUI:  eui64Ptr(types.EUI64{0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}),
			DevEUI:   eui64Ptr(types.EUI64{0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}),
		},
		ProvisionerID: "microchip",
		ProvisioningData: &pbtypes.Struct{
			Fields: map[string]*pbtypes.Value{
				"partNumber": {
					Kind: &pbtypes.Value_StringValue{
						StringValue: "test",
					},
				},
				"uniqueId": {
					Kind: &pbtypes.Value_StringValue{
						StringValue: "012302030405060727",
					},
				},
			},
		},
	}

	nwkKey, err := svc.GetNwkKey(ctx, dev)
	a.So(err, should.BeNil)
	a.So(nwkKey, should.Resemble, &types.AES128Key{0x84, 0xff, 0xb1, 0x28, 0xf4, 0x5d, 0xba, 0x1f, 0xa3, 0x67, 0x7f, 0x75, 0x2b, 0x4d, 0x93, 0x23})

	appKey, err := svc.GetAppKey(ctx, dev)
	a.So(err, should.BeNil)
	a.So(appKey, should.Resemble, &types.AES128Key{0x49, 0x0d, 0xf6, 0xe6, 0xf4, 0xc4, 0xe9, 0xa4, 0x33, 0x7e, 0x47, 0x4b, 0xf4, 0x73, 0x54, 0xde})
}
