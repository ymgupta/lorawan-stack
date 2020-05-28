// Copyright © 2019 The Things Industries B.V.

package commands

import (
	"context"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/prometheus/client_golang/prometheus"
	pkglicense "go.thethings.network/lorawan-stack/v3/pkg/license"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

var license *ttipb.License

func initializeLicense(ctx context.Context) (context.Context, error) {
	var err error
	license, err = pkglicense.Read(config.License)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	if license != nil {
		logger.WithFields(license).Info("Valid license")
	} else {
		now := time.Now()
		license = &ttipb.License{
			LicenseIdentifiers:      ttipb.LicenseIdentifiers{LicenseID: "unlicensed"},
			CreatedAt:               now,
			ValidFrom:               now,
			WarnFor:                 time.Hour,
			ValidUntil:              now.Add(time.Hour),
			ComponentAddressRegexps: []string{"localhost"},
			DevAddrPrefixes: []types.DevAddrPrefix{
				{DevAddr: types.DevAddr{0, 0, 0, 0}, Length: 7},
				{DevAddr: types.DevAddr{2, 0, 0, 0}, Length: 7},
			},
			MaxApplications:  &pbtypes.UInt64Value{Value: 10},
			MaxClients:       &pbtypes.UInt64Value{Value: 10},
			MaxEndDevices:    &pbtypes.UInt64Value{Value: 10},
			MaxGateways:      &pbtypes.UInt64Value{Value: 10},
			MaxOrganizations: &pbtypes.UInt64Value{Value: 10},
			MaxUsers:         &pbtypes.UInt64Value{Value: 10},
			Metering: &ttipb.MeteringConfiguration{
				Interval: time.Minute,
				Metering: &ttipb.MeteringConfiguration_Prometheus_{
					Prometheus: &ttipb.MeteringConfiguration_Prometheus{},
				},
			},
		}
		logger.WithFields(license).Warn("No license configured, running unlicensed mode")
	}
	ctx = pkglicense.NewContextWithLicense(ctx, *license)
	labels := prometheus.Labels{"metering_interval": "disabled"}
	if license.Metering != nil {
		labels["metering_interval"] = license.Metering.Interval.String()
	}
	metrics.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace:   metrics.Namespace,
		Name:        "license_expiry_seconds",
		Help:        "Expiry date of the license.",
		ConstLabels: labels,
	}, func() float64 {
		return float64(pkglicense.FromContext(ctx).ValidUntil.UnixNano()) / 1e9
	}))
	return ctx, nil
}
