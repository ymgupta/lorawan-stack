// Copyright © 2019 The Things Industries B.V.

package ttipb

import ttnpb "go.thethings.network/lorawan-stack/v3/pkg/ttnpb"

type randyIdentifiers interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

func NewPopulatedTenantIdentifiers(r randyIdentifiers, _ bool) *TenantIdentifiers {
	return &TenantIdentifiers{
		TenantID: ttnpb.NewPopulatedID(r),
	}
}

func NewPopulatedLicenseIssuerIdentifiers(r randyIdentifiers, _ bool) *LicenseIssuerIdentifiers {
	return &LicenseIssuerIdentifiers{
		LicenseIssuerID: ttnpb.NewPopulatedID(r),
	}
}

func NewPopulatedLicenseIdentifiers(r randyIdentifiers, _ bool) *LicenseIdentifiers {
	return &LicenseIdentifiers{
		LicenseID: ttnpb.NewPopulatedID(r),
	}
}
