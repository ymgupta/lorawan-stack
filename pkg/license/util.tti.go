// Copyright © 2019 The Things Industries B.V.

package license

import (
	"crypto"
	"crypto/ecdsa"
	"encoding/asn1"
	"errors"
	"math/big"
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
