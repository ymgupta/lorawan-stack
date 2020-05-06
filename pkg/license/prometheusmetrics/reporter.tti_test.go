// Copyright © 2019 The Things Industries B.V.

package prometheusmetrics_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/license/prometheusmetrics"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

func TestReporter(t *testing.T) {
	a := assertions.New(t)

	reg := prometheus.NewRegistry()
	r, err := prometheusmetrics.New(&ttipb.MeteringConfiguration_Prometheus{}, reg)
	a.So(err, should.BeNil)

	err = r.Report(test.Context(), &ttipb.MeteringData{
		Tenants: []*ttipb.MeteringData_TenantMeteringData{
			{
				TenantIdentifiers: ttipb.TenantIdentifiers{TenantID: "foo-tenant"},
				Totals: &ttipb.TenantRegistryTotals{
					Applications:  1,
					Clients:       2,
					EndDevices:    3,
					Gateways:      4,
					Organizations: 5,
					Users:         6,
				},
			},
		},
	})
	a.So(err, should.BeNil)

	metrics, _ := reg.Gather()
	for _, metric := range metrics {
		var expectedValue float64
		switch metric.GetName() {
		case "ttn_lw_tenant_metrics_applications":
			expectedValue = 1
		case "ttn_lw_tenant_metrics_clients":
			expectedValue = 2
		case "ttn_lw_tenant_metrics_end_devices":
			expectedValue = 3
		case "ttn_lw_tenant_metrics_gateways":
			expectedValue = 4
		case "ttn_lw_tenant_metrics_organizations":
			expectedValue = 5
		case "ttn_lw_tenant_metrics_users":
			expectedValue = 6
		default:
			t.Errorf("Unexpected metric: %q", metric.GetName())
		}
		if a.So(metric.GetMetric(), should.HaveLength, 1) {
			var foundTenantIDLabel bool
			for _, label := range metric.GetMetric()[0].GetLabel() {
				switch label.GetName() {
				case "tenant_id":
					foundTenantIDLabel = true
					a.So(label.GetValue(), should.Equal, "foo-tenant")
				default:
					t.Errorf("Unexpected metric label: %q", label.GetName())
				}
			}
			a.So(foundTenantIDLabel, should.BeTrue)
			a.So(metric.GetMetric()[0].GetGauge().GetValue(), should.Equal, expectedValue)
		}
	}
}
