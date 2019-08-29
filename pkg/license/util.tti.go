// Copyright © 2019 The Things Industries B.V.

package license

import (
	"crypto"
	"crypto/ecdsa"
	"encoding/asn1"
	"math/big"
	"time"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

var (
	errLicenseNotValidYet = errors.DefineFailedPrecondition("license_not_valid_yet", "the license is not valid yet", "valid_from")
	errLicenseExpired     = errors.DefineFailedPrecondition("license_expired", "the license is expired", "valid_until")
)

// CheckValidity checks the validity of the license.
func CheckValidity(license *ttipb.License) error {
	now := time.Now()
	if validFrom := license.GetValidFrom(); now.Before(validFrom) {
		return errLicenseNotValidYet.WithAttributes("valid_from", validFrom)
	}
	if validUntil := license.GetValidUntil(); now.After(validUntil) && (license.Metering == nil || !validUntil.IsZero()) {
		return errLicenseExpired.WithAttributes("valid_until", validUntil)
	}
	return nil
}

var (
	errUnknownLicenseKeyType = errors.DefineFailedPrecondition("unknown_license_key_type", "unknown license key type")
)

func getHash(pub crypto.PublicKey) (crypto.Hash, error) {
	switch pub := pub.(type) {
	case *ecdsa.PublicKey:
		switch pub.Curve.Params().BitSize {
		case 256:
			return crypto.SHA256, nil
		case 384:
			return crypto.SHA384, nil
		case 521:
			return crypto.SHA512, nil
		default:
			return 0, errUnknownLicenseKeyType
		}
	default:
		return 0, errUnknownLicenseKeyType
	}
}

type ecdsaSignature struct {
	R, S *big.Int
}

func (s ecdsaSignature) MarshalBinary() ([]byte, error) {
	return asn1.Marshal(s)
}

func (s *ecdsaSignature) UnmarshalBinary(b []byte) error {
	_, err := asn1.Unmarshal(b, s)
	return err
}
