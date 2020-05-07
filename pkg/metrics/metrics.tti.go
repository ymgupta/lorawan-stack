// Copyright © 2019 The Things Industries B.V.

//+build tti

package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
)

// ContextLabelNames are the label names that can be retrieved from a context for XXXVec metrics.
var ContextLabelNames = []string{"tenant_id"}

// LabelsFromContext returns the values for ContextLabelNames.
var LabelsFromContext = func(ctx context.Context) prometheus.Labels {
	return map[string]string{"tenant_id": tenant.FromContext(ctx).TenantID}
}
