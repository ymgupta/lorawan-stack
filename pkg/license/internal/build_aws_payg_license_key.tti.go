// Copyright © 2020 The Things Industries B.V.

//+build ignore

package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

var encoding = base64.StdEncoding

func main() {

	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	extensionDuration := 2 * time.Hour

	license := ttipb.License{
		LicenseIdentifiers:       ttipb.LicenseIdentifiers{LicenseID: "aws-marketplace-0001"},
		LicenseIssuerIdentifiers: ttipb.LicenseIssuerIdentifiers{LicenseIssuerID: "tti"},
		CreatedAt:                now,
		ValidFrom:                now,
		WarnFor:                  time.Hour,
		LimitFor:                 time.Hour,
		Components:               nil, // unrestricted
		DevAddrPrefixes:          []types.DevAddrPrefix{},
		JoinEUIPrefixes:          []types.EUI64Prefix{},
		MultiTenancy:             false,
		MaxApplications:          nil, // unrestricted
		MaxClients:               nil, // unrestricted
		MaxEndDevices:            nil, // unrestricted
		MaxGateways:              nil, // unrestricted
		MaxOrganizations:         nil, // unrestricted
		MaxUsers:                 nil, // unrestricted
		Metering: &ttipb.MeteringConfiguration{
			OnSuccess: &ttipb.LicenseUpdate{
				ExtendValidUntil: &extensionDuration,
			},
			Metering: &ttipb.MeteringConfiguration_AWS_{
				AWS: &ttipb.MeteringConfiguration_AWS{
					SKU: "b3xl44trs3xc6vgx904sb02qe",
				},
			},
		},
	}

	licenseKey, err := license.BuildLicenseKey()
	if err != nil {
		log.Fatal(err)
	}

	licenseKeyBytes, err := licenseKey.Marshal()
	if err != nil {
		log.Fatal(err)
	}

	encodedLicenseKeyBytes := make([]byte, encoding.EncodedLen(len(licenseKeyBytes)))
	encoding.Encode(encodedLicenseKeyBytes, licenseKeyBytes)

	os.Stdout.Write(encodedLicenseKeyBytes)
	fmt.Fprintln(os.Stdout)
}
