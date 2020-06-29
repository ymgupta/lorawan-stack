// Copyright © 2020 The Things Industries B.V.

//+build ignore

package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

var encoding = base64.StdEncoding

var defaultRoles = [...]ttnpb.ClusterRole{
	ttnpb.ClusterRole_NONE,
	ttnpb.ClusterRole_ENTITY_REGISTRY,
	ttnpb.ClusterRole_ACCESS,
	ttnpb.ClusterRole_GATEWAY_SERVER,
	ttnpb.ClusterRole_NETWORK_SERVER,
	ttnpb.ClusterRole_APPLICATION_SERVER,
	ttnpb.ClusterRole_JOIN_SERVER,
	ttnpb.ClusterRole_DEVICE_TEMPLATE_CONVERTER,
	ttnpb.ClusterRole_GATEWAY_CONFIGURATION_SERVER,
	ttnpb.ClusterRole_QR_CODE_GENERATOR,
	ttnpb.ClusterRole_PACKET_BROKER_AGENT,
}

var additionalRoles = [...]ttnpb.ClusterRole{
	ttnpb.ClusterRole_CRYPTO_SERVER,
	ttnpb.ClusterRole_DEVICE_CLAIMING_SERVER,
	ttnpb.ClusterRole_TENANT_BILLING_SERVER,
}

var (
	id            = pflag.String("id", "", "License ID")
	days          = pflag.Int("days", 180, "License expiry")
	warn          = pflag.Duration("warn", 24*time.Hour, "Warn before expiry")
	limit         = pflag.Duration("limit", 8*time.Hour, "Limit before expiry")
	allComponents = pflag.Bool("all-components", false, "License all available components")
	cs            = pflag.Bool("crypto-server", false, "License Crypto Server")
	dcs           = pflag.Bool("device-claiming-server", false, "License Device Claiming Server")
	tbs           = pflag.Bool("tenant-billing-server", false, "License Tenant Billing Server")
	multiTenancy  = pflag.Bool("multi-tenancy", false, "License multi-tenancy")

	addressRegexps  = pflag.StringSlice("address-regexps", nil, "Component address regexps")
	devAddrPrefixes = pflag.StringSlice("dev-addr-prefixes", nil, "DevAddr prefixes")
	joinEUIPrefixes = pflag.StringSlice("join-eui-prefixes", nil, "JoinEUI prefixes")

	applications  = pflag.Int("applications", 10, "Maximum number of applications")
	clients       = pflag.Int("clients", 10, "Maximum number of clients")
	endDevices    = pflag.Int("end-devices", 100, "Maximum number of end devices")
	gateways      = pflag.Int("gateways", 10, "Maximum number of gateways")
	organizations = pflag.Int("organizations", 10, "Maximum number of organizations")
	users         = pflag.Int("users", 10, "Maximum number of users")
)

func main() {
	pflag.Parse()

	if *id == "" {
		log.Fatal("Missing license ID")
	}

	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	until := now.Add(time.Duration(*days) * 24 * time.Hour)
	until = time.Date(until.Year(), until.Month(), until.Day(), 23, 59, 59, 0, time.UTC)

	components := defaultRoles[:]
	if *allComponents {
		for _, role := range additionalRoles {
			components = append(components, role)
		}
	} else if *cs {
		components = append(components, ttnpb.ClusterRole_CRYPTO_SERVER)
	} else if *dcs {
		components = append(components, ttnpb.ClusterRole_DEVICE_CLAIMING_SERVER)
	} else if *tbs {
		components = append(components, ttnpb.ClusterRole_TENANT_BILLING_SERVER)
	}

	for _, re := range *addressRegexps {
		if _, err := regexp.Compile(re); err != nil {
			log.Fatal(err)
		}
	}

	license := ttipb.License{
		LicenseIdentifiers:       ttipb.LicenseIdentifiers{LicenseID: *id},
		LicenseIssuerIdentifiers: ttipb.LicenseIssuerIdentifiers{LicenseIssuerID: "tti"},
		CreatedAt:                now,
		ValidFrom:                now,
		ValidUntil:               until,
		WarnFor:                  *warn,
		LimitFor:                 *limit,
		Components:               components,
		ComponentAddressRegexps:  *addressRegexps,
		MultiTenancy:             *multiTenancy,
	}

	for _, prefix := range *devAddrPrefixes {
		var p types.DevAddrPrefix
		if err := p.UnmarshalText([]byte(prefix)); err != nil {
			log.Fatal(err)
		}
		license.DevAddrPrefixes = append(license.DevAddrPrefixes, p)
	}

	for _, prefix := range *joinEUIPrefixes {
		var p types.EUI64Prefix
		if err := p.UnmarshalText([]byte(prefix)); err != nil {
			log.Fatal(err)
		}
		license.JoinEUIPrefixes = append(license.JoinEUIPrefixes, p)
	}

	if *applications > 0 {
		license.MaxApplications = &pbtypes.UInt64Value{Value: uint64(*applications)}
	}
	if *clients > 0 {
		license.MaxClients = &pbtypes.UInt64Value{Value: uint64(*clients)}
	}
	if *endDevices > 0 {
		license.MaxEndDevices = &pbtypes.UInt64Value{Value: uint64(*endDevices)}
	}
	if *gateways > 0 {
		license.MaxGateways = &pbtypes.UInt64Value{Value: uint64(*gateways)}
	}
	if *organizations > 0 {
		license.MaxOrganizations = &pbtypes.UInt64Value{Value: uint64(*organizations)}
	}
	if *users > 0 {
		license.MaxUsers = &pbtypes.UInt64Value{Value: uint64(*users)}
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
