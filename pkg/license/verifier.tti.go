// Copyright © 2019 The Things Industries B.V.

package license

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
)

type SignatureVerifier func(license []byte, sig []byte) error

var errInvalidLicenseSignature = errors.DefineFailedPrecondition("invalid_license_signature", "invalid license signature")

func newECDSAVerifier(hash crypto.Hash, pub *ecdsa.PublicKey) SignatureVerifier {
	return func(license []byte, sig []byte) error {
		var es ecdsaSignature
		if err := es.UnmarshalBinary(sig); err != nil {
			return err
		}
		h := hash.New()
		h.Write(license)
		if !ecdsa.Verify(pub, h.Sum(nil), es.R, es.S) {
			return errInvalidLicenseSignature.New()
		}
		return nil
	}
}

var verifiers = map[string]SignatureVerifier{}

var (
	errMissingLicense   = errors.DefineFailedPrecondition("missing_license", "missing license")
	errNoValidSignature = errors.DefineFailedPrecondition("no_valid_license_signature", "no valid license signature")
)

// VerifyKey verifies the license key and extracts license information.
func VerifyKey(licenseKey *ttipb.LicenseKey) (ttipb.License, error) {
	license, err := licenseKey.UnmarshalLicense()
	if err != nil {
		return ttipb.License{}, errInvalidLicense.WithCause(err)
	}
	if license == nil {
		return ttipb.License{}, errMissingLicense.New()
	}
	if err := CheckValidity(license); err != nil {
		return ttipb.License{}, err
	}
	var anyValid bool
	for _, sig := range licenseKey.GetSignatures() {
		if verify, ok := verifiers[sig.GetKeyID()]; ok {
			if err := verify(licenseKey.GetLicense(), sig.GetSignature()); err != nil {
				return ttipb.License{}, err
			}
			anyValid = true
		}
	}
	if !anyValid {
		return ttipb.License{}, errNoValidSignature.New()
	}
	return *license, nil
}

// NewVerifier returns a new license key verifier for the given public key.
func NewVerifier(public []byte) (SignatureVerifier, error) {
	pub, err := x509.ParsePKIXPublicKey(public)
	if err != nil {
		return nil, err
	}
	switch pub := pub.(type) {
	case *ecdsa.PublicKey:
		hash, err := getHash(pub)
		if err != nil {
			return nil, err
		}
		return newECDSAVerifier(hash, pub), nil
	default:
		return nil, errUnknownLicenseKeyType.New()
	}
}

// MustRegisterKey adds a verifier from the public key.
func MustRegisterKey(id string, pub []byte) {
	kv, err := NewVerifier(pub)
	if err != nil {
		panic(err)
	}
	verifiers[id] = kv
}
