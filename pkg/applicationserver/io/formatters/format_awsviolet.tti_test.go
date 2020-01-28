// Copyright © 2020 The Things Industries B.V.

package formatters

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

func TestMarshalAWSVioletDevEUI(t *testing.T) {
	for _, tt := range []struct {
		EUI    types.EUI64
		Expect string
	}{
		{types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}, "01-02-03-04-05-06-07-08"},
	} {
		t.Run(tt.EUI.String(), func(t *testing.T) {
			a := assertions.New(t)
			b, err := awsVioletDevEUI(tt.EUI).MarshalText()
			a.So(err, should.BeNil)
			a.So(string(b), should.Equal, tt.Expect)
		})
	}
}

func TestMarshalAWSVioletTimestamp(t *testing.T) {
	for _, tt := range []struct {
		Time   time.Time
		Expect string
	}{
		{time.Date(2020, time.January, 10, 22, 18, 41, 173669000, time.UTC), "1578694721.173669"},
	} {
		t.Run(tt.Time.Format(time.RFC3339Nano), func(t *testing.T) {
			a := assertions.New(t)
			b, err := awsVioletTimestamp(tt.Time).MarshalJSON()
			a.So(err, should.BeNil)
			a.So(string(b), should.Equal, tt.Expect)
		})
	}
}

func TestMarshalAWSVioletGatewayEUI(t *testing.T) {
	for _, tt := range []struct {
		EUI    types.EUI64
		Expect string
	}{
		{types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}, "72623859790382856"},
	} {
		t.Run(tt.EUI.String(), func(t *testing.T) {
			a := assertions.New(t)
			b, err := awsVioletGatewayEUI(tt.EUI).MarshalJSON()
			a.So(err, should.BeNil)
			a.So(string(b), should.Equal, tt.Expect)
		})
	}
}

func TestMarshalAWSVioletDevAddr(t *testing.T) {
	for _, tt := range []struct {
		Addr   types.DevAddr
		Expect string
	}{
		{types.DevAddr{1, 2, 3, 4}, "16909060"},
	} {
		t.Run(tt.Addr.String(), func(t *testing.T) {
			a := assertions.New(t)
			b, err := awsVioletDevAddr(tt.Addr).MarshalJSON()
			a.So(err, should.BeNil)
			a.So(string(b), should.Equal, tt.Expect)
		})
	}
}

func TestMarshalAWSVioletKeyEnvelope(t *testing.T) {
	k := types.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	wrappedKey, _ := crypto.WrapKey(k[:], k[:])
	kekID := sha256.Sum256(k[:])
	kekLabel := hex.EncodeToString(kekID[:])

	for _, tt := range []struct {
		Env    ttnpb.KeyEnvelope
		Expect string
	}{
		{ttnpb.KeyEnvelope{
			KEKLabel: "plain",
			Key:      &k,
		}, "plain:01020304050607080102030405060708"},
		{ttnpb.KeyEnvelope{
			KEKLabel:     kekLabel,
			EncryptedKey: wrappedKey,
		}, "aes:AIyNVk+HP6GIrAeGCnvS87zgmfpuyHXA4Lk2Nu6GlgS+cOoUA18TRDPg/3O9K0na86E8qscwzpy4"},
	} {
		t.Run(tt.Env.KEKLabel, func(t *testing.T) {
			a := assertions.New(t)
			b, err := awsVioletKeyEnvelope(tt.Env).MarshalText()
			a.So(err, should.BeNil)
			a.So(string(b), should.Equal, tt.Expect)
		})
	}
}

func TestMarshalAWSVioletPayload(t *testing.T) {
	for _, tt := range []struct {
		Payload []byte
		Expect  string
	}{
		{[]byte{1, 2, 3, 4}, "01020304"},
	} {
		t.Run(fmt.Sprintf("%s", tt.Payload), func(t *testing.T) {
			a := assertions.New(t)
			b, err := awsVioletPayload(tt.Payload).MarshalText()
			a.So(err, should.BeNil)
			a.So(string(b), should.Equal, tt.Expect)
		})
	}
}
