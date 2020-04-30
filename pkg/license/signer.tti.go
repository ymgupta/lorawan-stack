// Copyright © 2019 The Things Industries B.V.

package license

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

// KeySigner verifies the license key and extracts license information.
type KeySigner interface {
	SignKey(licenseKey *ttipb.LicenseKey) error
}

type keySigner struct {
	keyID string
	sign  func(license []byte) (sig []byte, err error)
}

func (ks *keySigner) SignKey(licenseKey *ttipb.LicenseKey) error {
	var err error
	sig, err := ks.sign(licenseKey.License)
	if err != nil {
		return err
	}
	licenseKey.Signatures = append(licenseKey.Signatures, &ttipb.LicenseKey_Signature{
		KeyID:     ks.keyID,
		Signature: sig,
	})
	return nil
}

func newECDSASigner(hash crypto.Hash, priv *ecdsa.PrivateKey) func(license []byte) (sig []byte, err error) {
	return func(license []byte) (sig []byte, err error) {
		h := hash.New()
		h.Write(license)
		r, s, err := ecdsa.Sign(rand.Reader, priv, h.Sum(nil))
		if err != nil {
			return nil, err
		}
		return ecdsaSignature{R: r, S: s}.MarshalBinary()
	}
}

// NewSigner returns a new license key signer for the given private key.
func NewSigner(keyID, keyType string, private []byte) (KeySigner, error) {
	switch strings.TrimSuffix(strings.ToLower(keyType), " private key") {
	case "ec":
		priv, err := x509.ParseECPrivateKey(private)
		if err != nil {
			return nil, err
		}
		hash, err := getHash(priv.Public())
		if err != nil {
			return nil, err
		}
		return &keySigner{keyID: keyID, sign: newECDSASigner(hash, priv)}, nil
	default:
		return nil, errUnknownLicenseKeyType.New()
	}
}
