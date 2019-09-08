// Copyright © 2019 The Things Industries B.V.

package deviceclaimingserver

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/metrics"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtClaimEndDeviceSuccess = events.Define(
		"dcs.end_device.claim.success", "claim end device successful",
		ttnpb.RIGHT_APPLICATION_DEVICES_READ,
	)
	evtClaimEndDeviceFailure = events.Define(
		"dcs.end_device.claim.fail", "claim end device failure",
		ttnpb.RIGHT_APPLICATION_DEVICES_READ,
	)
)

const (
	subsystem     = "dcs"
	unknown       = "unknown"
	applicationID = "application_id"
)

var dcsMetrics = &claimMetrics{
	endDevicesClaimSucceeded: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "claim_end_devices_success_total",
			Help:      "Total number of successfully claimed end devices",
		},
		[]string{applicationID},
	),
	endDevicesClaimFailed: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "claim_end_device_failed_total",
			Help:      "Total number of claim end devices failures",
		},
		[]string{applicationID, "error"},
	),
}

func init() {
	metrics.MustRegister(dcsMetrics)
}

type claimMetrics struct {
	endDevicesClaimSucceeded *metrics.ContextualCounterVec
	endDevicesClaimFailed    *metrics.ContextualCounterVec
}

func (m claimMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.endDevicesClaimSucceeded.Describe(ch)
	m.endDevicesClaimFailed.Describe(ch)
}

func (m claimMetrics) Collect(ch chan<- prometheus.Metric) {
	m.endDevicesClaimSucceeded.Collect(ch)
	m.endDevicesClaimFailed.Collect(ch)
}

func registerSuccessClaimEndDevice(ctx context.Context, ids ttnpb.EndDeviceIdentifiers) {
	events.Publish(evtClaimEndDeviceSuccess(ctx, ids, nil))
	dcsMetrics.endDevicesClaimSucceeded.WithLabelValues(ctx, ids.ApplicationID).Inc()
}

func registerFailClaimEndDevice(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, err error) {
	events.Publish(evtClaimEndDeviceFailure(ctx, ids, nil))
	if ttnErr, ok := errors.From(err); ok {
		dcsMetrics.endDevicesClaimFailed.WithLabelValues(ctx, ids.ApplicationID, ttnErr.FullName()).Inc()
	} else {
		dcsMetrics.endDevicesClaimFailed.WithLabelValues(ctx, ids.ApplicationID, unknown).Inc()
	}
}
