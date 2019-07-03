// Copyright © 2019 The Things Industries B.V.

package microchip

import (
	"bytes"
	"crypto/sha256"

	"go.thethings.network/lorawan-stack/pkg/types"
)

// Key is a 256-bit secret key.
type Key [32]byte

// Challenge is a SHA-256 MAC challenge.
type Challenge [32]byte

// MAC is a SHA-256 digest.
type MAC [32]byte

// SerialNumber is an ATECC608A serial number.
type SerialNumber [9]byte

// DeriveKey derives a key from the given parent key, temporary key and serial number.
// This function resembles atcah_derive_key with mode DERIVE_KEY_RANDOM_FLAG.
func DeriveKey(targetKeyID uint16, parentKey, tempKey Key, sn SerialNumber) Key {
	mode := uint8(0x04)
	buf := make([]byte, 0, 96)

	// Block 1: Parent Key (32)
	buf = append(buf, parentKey[:]...)

	// Block 2: OpCode (1) | Mode (1) | Target Key ID (2) | SN[8] (1) | SN[0:1] (2) | pad32
	buf = append(buf, 0x1c)
	buf = append(buf, mode)
	buf = append(buf, uint8(targetKeyID), uint8(targetKeyID>>8))
	buf = append(buf, sn[8])
	buf = append(buf, sn[0], sn[1])
	buf = append(buf, bytes.Repeat([]byte{0x00}, 25)...)

	// Block 3: Temp Key (32)
	buf = append(buf, tempKey[:]...)

	return Key(sha256.Sum256(buf))
}

// KeyMAC generates an SHA-256 digest (MAC) of a key and challenge.
// This function resembles atcah_mac with mode 0.
func KeyMAC(keyID uint16, challenge Challenge, key Key, sn SerialNumber) MAC {
	mode := uint8(0x00)
	buf := make([]byte, 0, 88)

	// Block 1: Key (32)
	buf = append(buf, key[:]...)

	// Block 2: Challenge (32)
	buf = append(buf, challenge[:]...)

	// Block 3: OpCode (1) | Mode (1) | Key ID (2) | (11) | SN[8] (1) | (4) | SN[0:1] (2) | (2)
	buf = append(buf, 0x08)
	buf = append(buf, mode)
	buf = append(buf, uint8(keyID), uint8(keyID>>8))
	// See atcah_include_data with mode 0.
	buf = append(buf, bytes.Repeat([]byte{0x00}, 11)...)
	buf = append(buf, sn[8])
	buf = append(buf, bytes.Repeat([]byte{0x00}, 4)...)
	buf = append(buf, sn[0], sn[1])
	buf = append(buf, bytes.Repeat([]byte{0x00}, 2)...)

	return MAC(sha256.Sum256(buf))
}

// DiversifiedKey returns the 256-bit derived key from the given parent key and serial number.
// This function implements the Diversified Key Algorithm as defined by Microchip and The Things Industries.
func DiversifiedKey(parentKey Key, sn SerialNumber) Key {
	var tempKey Key
	copy(tempKey[0:], sn[0:4])
	copy(tempKey[4:], sn[4:9])
	return DeriveKey(0x0000, parentKey, tempKey, sn)
}

// DiversifiedKeyMAC returns an SHA-256 digest (MAC) of the diversified key of the given parent key and serial number.
// This function uses the Diversified Key Algorithm as defined by Microchip and The Things Industries.
func DiversifiedKeyMAC(parentKey Key, sn SerialNumber, challenge Challenge) MAC {
	key := DiversifiedKey(parentKey, sn)
	return KeyMAC(0x0000, challenge, key, sn)
}

// DiversifiedRootKeys derives the LoRaWAN root keys for the given serial number.
// This function uses the Diversified Key Algorithm as defined by Microchip and The Things Industries.
func DiversifiedRootKeys(parentKey Key, sn SerialNumber) (nwkKey, appKey types.AES128Key) {
	key := DiversifiedKey(parentKey, sn)
	copy(nwkKey[:], key[:16])
	copy(appKey[:], key[16:])
	return
}
