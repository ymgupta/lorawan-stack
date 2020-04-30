// Copyright © 2019 The Things Industries B.V.

//+build ignore

package main

import (
	"encoding/base64"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

var (
	signingKeyPath = flag.String("signing-key", "", "Path to signing key")
	signingKeyID   = flag.String("signing-key-id", "", "Signing key ID (default: filename of signing key without .pem)")
)

var encoding = base64.StdEncoding

// Usage:
//
// echo $LICENSE_KEY | go run ./pkg/license/internal/sign_license_key.tti.go -signing-key=/keybase/team/ttn.licensing/tti.lorawan-stack.v3.VERSION.pem
func main() {
	flag.Parse()

	if *signingKeyID == "" {
		*signingKeyID = strings.TrimSuffix(filepath.Base(*signingKeyPath), ".pem")
	}

	signingKeyBytes, err := ioutil.ReadFile(*signingKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	signingKeyBlock, _ := pem.Decode(signingKeyBytes)
	signingKeyType := strings.TrimSuffix(strings.ToLower(signingKeyBlock.Type), " private key")

	signer, err := license.NewSigner(*signingKeyID, signingKeyType, signingKeyBlock.Bytes)
	if err != nil {
		log.Fatal(err)
	}

	encodedLicenseKeyBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	licenseKeyBytes := make([]byte, encoding.DecodedLen(len(encodedLicenseKeyBytes)))
	n, err := encoding.Decode(licenseKeyBytes, encodedLicenseKeyBytes)
	if err != nil {
		log.Fatal(err)
	}
	licenseKeyBytes = licenseKeyBytes[:n]

	var licenseKey ttipb.LicenseKey
	if err = licenseKey.Unmarshal(licenseKeyBytes); err != nil {
		log.Fatal(err)
	}
	if err = signer.SignKey(&licenseKey); err != nil {
		log.Fatal(err)
	}

	licenseKeyBytes, err = licenseKey.Marshal()
	if err != nil {
		log.Fatal(err)
	}

	encodedLicenseKeyBytes = make([]byte, encoding.EncodedLen(len(licenseKeyBytes)))
	encoding.Encode(encodedLicenseKeyBytes, licenseKeyBytes)

	os.Stdout.Write(encodedLicenseKeyBytes)
	fmt.Fprintln(os.Stdout)
}
