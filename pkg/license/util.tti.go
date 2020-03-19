// Copyright © 2019 The Things Industries B.V.

package license

import (
	"crypto"
	"crypto/ecdsa"
	"encoding/asn1"
	"math/big"
	"strings"
	"time"

	"github.com/blang/semver"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/version"
)

var (
	errInvalidLicense     = errors.DefineFailedPrecondition("invalid_license", "invalid license")
	errLicenseNotValidYet = errors.DefineFailedPrecondition("license_not_valid_yet", "the license is not valid yet", "valid_from")
	errLicenseExpired     = errors.DefineFailedPrecondition("license_expired", "the license is expired", "valid_until")
	errLicenseNotRenewed  = errors.DefineFailedPrecondition("license_not_renewed", "the license was not renewed by the metering service")
	errVersionTooLow      = errors.DefineFailedPrecondition("version_too_low", "the current version number is too low", "min_version")
	errVersionTooHigh     = errors.DefineFailedPrecondition("version_too_high", "the current version number is too high", "max_version")
)

// CheckValidity checks the validity of the license.
func CheckValidity(license *ttipb.License) error {
	now := time.Now()
	if validFrom := license.GetValidFrom(); now.Before(validFrom) {
		return errLicenseNotValidYet.WithAttributes("valid_from", validFrom.Format(time.RFC822))
	}
	if validUntil := license.GetValidUntil(); !validUntil.IsZero() && now.After(validUntil) {
		if license.Metering != nil {
			return errLicenseNotRenewed.New()
		}
		return errLicenseExpired.WithAttributes("valid_until", validUntil.Format(time.RFC822))
	}
	currentVersion, _ := semver.Parse(strings.TrimPrefix(version.TTN, "v")) // Invalid versions (snapshots) are 0.0.0.
	if minVersionStr := license.GetMinVersion(); minVersionStr != "" {
		if minVersion, err := semver.Parse(minVersionStr); err == nil && currentVersion.Compare(minVersion) < 0 {
			return errVersionTooLow.WithAttributes("min_version", minVersionStr)
		}
	}
	if maxVersionStr := license.GetMaxVersion(); maxVersionStr != "" {
		if maxVersion, err := semver.Parse(maxVersionStr); err == nil && currentVersion.Compare(maxVersion) > 0 {
			return errVersionTooHigh.WithAttributes("max_version", maxVersionStr)
		}
	}
	return nil
}

var errLimitedFunctionality = errors.DefineFailedPrecondition("limited_functionality", "limited functionality due to license expiry")

// CheckLimitedFunctionality checks if functionality needs to be limited.
func CheckLimitedFunctionality(license *ttipb.License) error {
	if validUntil := license.GetValidUntil(); !validUntil.IsZero() && time.Until(validUntil) < license.GetLimitFor() {
		return errLimitedFunctionality.New()
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
