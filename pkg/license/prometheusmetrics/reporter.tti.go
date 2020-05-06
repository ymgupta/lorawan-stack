// Copyright © 2019 The Things Industries B.V.

package prometheusmetrics

import (
	"context"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
)

type tenantMetrics struct {
	sync.RWMutex
	applications  *prometheus.GaugeVec
	clients       *prometheus.GaugeVec
	endDevices    *prometheus.GaugeVec
	gateways      *prometheus.GaugeVec
	organizations *prometheus.GaugeVec
	users         *prometheus.GaugeVec
}

func (m *tenantMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.RLock()
	defer m.RUnlock()
	m.applications.Describe(ch)
	m.clients.Describe(ch)
	m.endDevices.Describe(ch)
	m.gateways.Describe(ch)
	m.organizations.Describe(ch)
	m.users.Describe(ch)
}

func (m *tenantMetrics) Collect(ch chan<- prometheus.Metric) {
	m.RLock()
	defer m.RUnlock()
	m.applications.Collect(ch)
	m.clients.Collect(ch)
	m.endDevices.Collect(ch)
	m.gateways.Collect(ch)
	m.organizations.Collect(ch)
	m.users.Collect(ch)
}

const (
	subsystem = "tenant_metrics"
	tenantID  = "tenant_id"
)

func newTenantMetrics() *tenantMetrics {
	return &tenantMetrics{
		applications: metrics.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      "applications",
				Help:      "Number of registered applications",
			},
			[]string{tenantID},
		),
		clients: metrics.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      "clients",
				Help:      "Number of registered clients",
			},
			[]string{tenantID},
		),
		endDevices: metrics.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      "end_devices",
				Help:      "Number of registered end_devices",
			},
			[]string{tenantID},
		),
		gateways: metrics.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      "gateways",
				Help:      "Number of registered gateways",
			},
			[]string{tenantID},
		),
		organizations: metrics.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      "organizations",
				Help:      "Number of registered organizations",
			},
			[]string{tenantID},
		),
		users: metrics.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      "users",
				Help:      "Number of registered users",
			},
			[]string{tenantID},
		),
	}
}

// Reporter reports tenant metrics to Prometheus.
type Reporter struct {
	*tenantMetrics
}

// Report implements the license.MeteringReporter interface.
func (m *Reporter) Report(ctx context.Context, data *ttipb.MeteringData) error {
	m.tenantMetrics.Lock()
	defer m.tenantMetrics.Unlock()
	m.tenantMetrics.applications.Reset()
	m.tenantMetrics.clients.Reset()
	m.tenantMetrics.endDevices.Reset()
	m.tenantMetrics.gateways.Reset()
	m.tenantMetrics.organizations.Reset()
	m.tenantMetrics.users.Reset()
	for _, data := range data.Tenants {
		m.tenantMetrics.applications.
			WithLabelValues(data.GetTenantID()).
			Set(float64(data.GetTotals().GetApplications()))
		m.tenantMetrics.clients.
			WithLabelValues(data.GetTenantID()).
			Set(float64(data.GetTotals().GetClients()))
		m.tenantMetrics.endDevices.
			WithLabelValues(data.GetTenantID()).
			Set(float64(data.GetTotals().GetEndDevices()))
		m.tenantMetrics.gateways.
			WithLabelValues(data.GetTenantID()).
			Set(float64(data.GetTotals().GetGateways()))
		m.tenantMetrics.organizations.
			WithLabelValues(data.GetTenantID()).
			Set(float64(data.GetTotals().GetOrganizations()))
		m.tenantMetrics.users.
			WithLabelValues(data.GetTenantID()).
			Set(float64(data.GetTotals().GetUsers()))
	}
	return nil
}

// New returns a new tenant metrics reporter that exports metrics to Prometheus.
func New(config *ttipb.MeteringConfiguration_Prometheus, reg prometheus.Registerer) (*Reporter, error) {
	m := &Reporter{
		tenantMetrics: newTenantMetrics(),
	}
	if err := reg.Register(m); err != nil {
		return nil, err
	}
	return m, nil
}
