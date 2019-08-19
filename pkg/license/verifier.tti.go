// Copyright © 2019 The Things Industries B.V.

package license

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	"errors"
	"fmt"
	"time"

	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

type SignatureVerifier func(license []byte, sig []byte) error

func newECDSAVerifier(hash crypto.Hash, pub *ecdsa.PublicKey) SignatureVerifier {
	return func(license []byte, sig []byte) error {
		var es ecdsaSignature
		if err := es.UnmarshalBinary(sig); err != nil {
			return err
		}
		h := hash.New()
		h.Write(license)
		if !ecdsa.Verify(pub, h.Sum(nil), es.R, es.S) {
			return errors.New("invalid license signature")
		}
		return nil
	}
}

var verifiers = map[string]SignatureVerifier{}

// VerifyKey verifies the license key and extracts license information.
func VerifyKey(licenseKey *ttipb.LicenseKey) (ttipb.License, error) {
	license, err := licenseKey.UnmarshalLicense()
	if err != nil {
		return ttipb.License{}, err
	}
	if license == nil {
		return ttipb.License{}, errors.New("no license")
	}
	now := time.Now()
	if validFrom := license.GetValidFrom(); now.Before(validFrom) {
		return ttipb.License{}, fmt.Errorf("license is valid from %s", validFrom)
	}
	if validUntil := license.GetValidUntil(); now.After(validUntil) {
		return ttipb.License{}, fmt.Errorf("license is valid until %s", validUntil)
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
		return ttipb.License{}, errors.New("unknown license signature key ID")
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
		return nil, errors.New("unknown public license key type")
	}
}

func mustRegisterKey(id string, pub []byte) {
	kv, err := NewVerifier(pub)
	if err != nil {
		panic(err)
	}
	verifiers[id] = kv
}
