// Copyright © 2019 The Things Industries B.V.

//+build ignore

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
)

// Usage:
//
// go run ./pkg/license/internal/generate_signing_key.tti.go > /keybase/team/ttn.licensing/tti.lorawan-stack.v3.VERSION.pem
//
// Add the output of stderr to ./pkg/license/keys.tti.go.
func main() {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	privKeyBytes, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		log.Fatal(err)
	}
	err = pem.Encode(os.Stdout, &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privKeyBytes,
	})
	if err != nil {
		log.Fatal(err)
	}

	pubKey := privKey.Public()
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprint(os.Stderr, "mustRegisterKey(\"tti.lorawan-stack.v3.VERSION\", []byte{\n\t")
	for i, b := range pubKeyBytes {
		fmt.Fprintf(os.Stderr, "0x%02x,", b)
		switch {
		case i == len(pubKeyBytes)-1:
			fmt.Fprint(os.Stderr, "\n")
		case i%8 == 7:
			fmt.Fprint(os.Stderr, "\n\t")
		default:
			fmt.Fprint(os.Stderr, " ")
		}
	}
	fmt.Fprintln(os.Stderr, "})")
}
