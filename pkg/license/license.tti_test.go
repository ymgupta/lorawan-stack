// Copyright © 2019 The Things Industries B.V.

package license_test

import (
	"encoding/base64"
	"encoding/pem"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestLicense(t *testing.T) {
	a := assertions.New(t)

	privateKeyBlock, _ := pem.Decode([]byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEINgiCPlz0tPDnk/+MBnLEDTVI9So3vmC11vuogrndXwWoAoGCCqGSM49
AwEHoUQDQgAEbcZyAkH6QwLCd+rYqTYtIJ+cfn6K8cfJl+YaohBwSYSZxu/hCPOe
aRlnEgCYSEoIuptQQQCSZt1lelnqMwUw9A==
-----END EC PRIVATE KEY-----`))

	MustRegisterKey("TestKey", []byte{
		0x30, 0x59, 0x30, 0x13, 0x06, 0x07, 0x2a, 0x86,
		0x48, 0xce, 0x3d, 0x02, 0x01, 0x06, 0x08, 0x2a,
		0x86, 0x48, 0xce, 0x3d, 0x03, 0x01, 0x07, 0x03,
		0x42, 0x00, 0x04, 0x6d, 0xc6, 0x72, 0x02, 0x41,
		0xfa, 0x43, 0x02, 0xc2, 0x77, 0xea, 0xd8, 0xa9,
		0x36, 0x2d, 0x20, 0x9f, 0x9c, 0x7e, 0x7e, 0x8a,
		0xf1, 0xc7, 0xc9, 0x97, 0xe6, 0x1a, 0xa2, 0x10,
		0x70, 0x49, 0x84, 0x99, 0xc6, 0xef, 0xe1, 0x08,
		0xf3, 0x9e, 0x69, 0x19, 0x67, 0x12, 0x00, 0x98,
		0x48, 0x4a, 0x08, 0xba, 0x9b, 0x50, 0x41, 0x00,
		0x92, 0x66, 0xdd, 0x65, 0x7a, 0x59, 0xea, 0x33,
		0x05, 0x30, 0xf4,
	})

	s, err := NewSigner("TestKey", privateKeyBlock.Type, privateKeyBlock.Bytes)
	a.So(err, should.BeNil)

	now := time.Now().UTC()

	t.Run("Valid License", func(t *testing.T) {
		a := assertions.New(t)

		validLicense := &ttipb.License{
			LicenseIdentifiers: ttipb.LicenseIdentifiers{LicenseID: "test-valid"},
			CreatedAt:          now.Add(-24 * time.Hour),
			ValidFrom:          now.Add(-1 * time.Hour),
			ValidUntil:         now.Add(time.Hour),
		}

		validKey, err := validLicense.BuildLicenseKey()
		a.So(err, should.BeNil)

		err = s.SignKey(validKey)
		a.So(err, should.BeNil)

		res, err := VerifyKey(validKey)
		a.So(err, should.BeNil)
		a.So(res, should.Resemble, *validLicense)

		validKeyBytes, err := validKey.Marshal()
		if err != nil {
			t.Fatal(err)
		}

		fromConfig, err := Read(Config{
			Key: base64.StdEncoding.EncodeToString(validKeyBytes),
		})
		a.So(err, should.BeNil)
		a.So(fromConfig, should.Resemble, validLicense)
	})

	t.Run("Not Yet Valid License", func(t *testing.T) {
		a := assertions.New(t)

		notYetValidKey, err := (&ttipb.License{
			LicenseIdentifiers: ttipb.LicenseIdentifiers{LicenseID: "test-not-yet-valid"},
			CreatedAt:          now.Add(-24 * time.Hour),
			ValidFrom:          now.Add(time.Hour),
			ValidUntil:         now.Add(2 * time.Hour),
		}).BuildLicenseKey()
		a.So(err, should.BeNil)

		err = s.SignKey(notYetValidKey)
		a.So(err, should.BeNil)

		_, err = VerifyKey(notYetValidKey)
		a.So(err, should.NotBeNil)
	})

	t.Run("Expired License", func(t *testing.T) {
		a := assertions.New(t)

		expiredKey, err := (&ttipb.License{
			LicenseIdentifiers: ttipb.LicenseIdentifiers{LicenseID: "test-expired"},
			CreatedAt:          now.Add(-24 * time.Hour),
			ValidFrom:          now.Add(-2 * time.Hour),
			ValidUntil:         now.Add(-1 * time.Hour),
		}).BuildLicenseKey()
		a.So(err, should.BeNil)

		err = s.SignKey(expiredKey)
		a.So(err, should.BeNil)

		_, err = VerifyKey(expiredKey)
		a.So(err, should.NotBeNil)
	})
}
