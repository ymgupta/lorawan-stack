// Copyright © 2019 The Things Industries B.V.

package license

import (
	"crypto"
	"crypto/ecdsa"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"
	"time"

	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

func checkValidity(license *ttipb.License) error {
	now := time.Now()
	if validFrom := license.GetValidFrom(); now.Before(validFrom) {
		return fmt.Errorf("license is valid from %s", validFrom)
	}
	if validUntil := license.GetValidUntil(); now.After(validUntil) {
		return fmt.Errorf("license is valid until %s", validUntil)
	}
	return nil
}

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
			return 0, errors.New("unsupported license key curve")
		}
	default:
		return 0, errors.New("unknown public license key type")
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
