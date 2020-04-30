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
	evtClaimEndDeviceAbort = events.Define(
		"dcs.end_device.claim.abort", "claim end device abort",
		ttnpb.RIGHT_APPLICATION_DEVICES_READ,
	)
	evtClaimEndDeviceFail = events.Define(
		"dcs.end_device.claim.fail", "claim end device fail",
		ttnpb.RIGHT_APPLICATION_DEVICES_READ,
		ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS,
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
	endDevicesClaimAborted: metrics.NewContextualCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "claim_end_device_aborted_total",
			Help:      "Total number of claim end devices abortions",
		},
		[]string{applicationID, "error"},
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
	endDevicesClaimAborted   *metrics.ContextualCounterVec
	endDevicesClaimFailed    *metrics.ContextualCounterVec
}

func (m claimMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.endDevicesClaimSucceeded.Describe(ch)
	m.endDevicesClaimAborted.Describe(ch)
	m.endDevicesClaimFailed.Describe(ch)
}

func (m claimMetrics) Collect(ch chan<- prometheus.Metric) {
	m.endDevicesClaimSucceeded.Collect(ch)
	m.endDevicesClaimAborted.Collect(ch)
	m.endDevicesClaimFailed.Collect(ch)
}

func registerSuccessClaimEndDevice(ctx context.Context, ids ttnpb.EndDeviceIdentifiers) {
	events.Publish(evtClaimEndDeviceSuccess(ctx, ids, nil))
	dcsMetrics.endDevicesClaimSucceeded.WithLabelValues(ctx, ids.ApplicationID).Inc()
}

func registerAbortClaimEndDevice(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, err error) {
	events.Publish(evtClaimEndDeviceAbort(ctx, ids, err))
	if ttnErr, ok := errors.From(err); ok {
		dcsMetrics.endDevicesClaimAborted.WithLabelValues(ctx, ids.ApplicationID, ttnErr.FullName()).Inc()
	} else {
		dcsMetrics.endDevicesClaimAborted.WithLabelValues(ctx, ids.ApplicationID, unknown).Inc()
	}
}

func registerFailClaimEndDevice(ctx context.Context, dev *ttnpb.EndDevice, err error) {
	events.Publish(evtClaimEndDeviceFail(ctx, dev.EndDeviceIdentifiers, dev))
	if ttnErr, ok := errors.From(err); ok {
		dcsMetrics.endDevicesClaimFailed.WithLabelValues(ctx, dev.ApplicationID, ttnErr.FullName()).Inc()
	} else {
		dcsMetrics.endDevicesClaimFailed.WithLabelValues(ctx, dev.ApplicationID, unknown).Inc()
	}
}
