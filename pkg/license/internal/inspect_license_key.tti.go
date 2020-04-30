// Copyright © 2019 The Things Industries B.V.

//+build ignore

package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"go.thethings.network/lorawan-stack/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

var encoding = base64.StdEncoding

// Usage:
//
// echo $LICENSE_KEY | go run ./pkg/license/internal/inspect_license_key.tti.go
func main() {
	flag.Parse()

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

	license, err := license.VerifyKey(&licenseKey)
	if err == nil {
		fmt.Fprintln(os.Stderr, "Verified License")
	} else {
		fmt.Fprintln(os.Stderr, "Unverified License")
		unverifiedLicense, err := licenseKey.UnmarshalLicense()
		if err != nil {
			log.Fatal(err)
		}
		license = *unverifiedLicense
	}

	json := jsonpb.TTN()
	json.Indent = "  "
	licenseJSON, err := json.Marshal(&license)
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(licenseJSON)

	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stderr, "Signatures:")
	for _, sig := range licenseKey.Signatures {
		fmt.Fprintf(os.Stderr, "- %s\n", sig.KeyID)
	}
}
