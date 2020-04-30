// Copyright © 2020 The Things Industries B.V.

//+build ignore

package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

var encoding = base64.StdEncoding

var (
	cluster = pflag.String("cluster", "eu1", "Cluster ID")
	billing = pflag.Bool("billing", false, "Integrate with billing")
)

func main() {
	pflag.Parse()

	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	until := now.Add(180 * 24 * time.Hour)
	until = time.Date(until.Year(), until.Month(), until.Day(), 23, 59, 59, 0, time.UTC)

	license := ttipb.License{
		LicenseIdentifiers: ttipb.LicenseIdentifiers{
			LicenseID: fmt.Sprintf("tti-production-%s", *cluster),
		},
		LicenseIssuerIdentifiers: ttipb.LicenseIssuerIdentifiers{LicenseIssuerID: "tti"},
		CreatedAt:                now,
		ValidFrom:                now,
		ValidUntil:               until,
		WarnFor:                  24 * time.Hour,
		LimitFor:                 8 * time.Hour,
		Components:               nil, // unrestricted
		ComponentAddressRegexps: []string{
			fmt.Sprintf(`^%s\.cloud\.thethings\.industries$`, *cluster),
		},
		DevAddrPrefixes:  []types.DevAddrPrefix{},
		JoinEUIPrefixes:  []types.EUI64Prefix{},
		MultiTenancy:     true,
		MaxApplications:  nil, // unrestricted
		MaxClients:       nil, // unrestricted
		MaxEndDevices:    nil, // unrestricted
		MaxGateways:      nil, // unrestricted
		MaxOrganizations: nil, // unrestricted
		MaxUsers:         nil, // unrestricted
	}

	if *billing {
		license.Metering = &ttipb.MeteringConfiguration{
			Metering: &ttipb.MeteringConfiguration_TenantBillingServer_{
				TenantBillingServer: &ttipb.MeteringConfiguration_TenantBillingServer{},
			},
		}
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
