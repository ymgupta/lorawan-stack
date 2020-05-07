// Copyright © 2019 The Things Industries B.V.

package license

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/license/awsmetrics"
	"go.thethings.network/lorawan-stack/v3/pkg/license/prometheusmetrics"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/metrics"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

// Cluster is the interface used for getting metrics.
type Cluster interface {
	Auth() grpc.CallOption
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole) (*grpc.ClientConn, error)
}

// MeteringReporter is the interface used for reporting metrics.
type MeteringReporter interface {
	Report(ctx context.Context, data *ttipb.MeteringData) error
}

type meteringSetup struct {
	config   *ttipb.MeteringConfiguration
	cluster  Cluster
	reporter MeteringReporter

	mu    sync.RWMutex                      // Protects the apply func.
	apply func(ttipb.License) ttipb.License // Applies the OnSuccess rules of the latest renewal to the license.
}

// Apply updates the license according to the update rules.
func (s *meteringSetup) Apply(license ttipb.License) ttipb.License {
	if s == nil {
		return license
	}
	s.mu.RLock()
	if s.apply != nil {
		license = s.apply(license)
	}
	s.mu.RUnlock()
	return license
}

// CollectAndReport collects metrics from the cluster and reports them to the MeteringReporter.
func (s *meteringSetup) CollectAndReport(ctx context.Context) (err error) {
	if s.reporter == nil {
		return errors.New("metering service reporter is not properly set up")
	}

	defer func() {
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Could not communicate with metering service.")
		}
	}()

	var meteringData ttipb.MeteringData

	cc, err := s.cluster.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY)
	if err != nil {
		return err
	}

	reg := ttipb.NewTenantRegistryClient(cc)
	creds := s.cluster.Auth()

	var (
		tenantsPageSize = 1000
		tenantsPage     = 1
	)
	for {
		res, err := reg.List(ctx, &ttipb.ListTenantsRequest{
			FieldMask: pbtypes.FieldMask{Paths: []string{"ids"}},
			Limit:     uint32(tenantsPageSize),
			Page:      uint32(tenantsPage),
		}, creds)
		if err != nil {
			return err
		}
		for _, tenant := range res.Tenants {
			totals, err := reg.GetRegistryTotals(ctx, &ttipb.GetTenantRegistryTotalsRequest{
				TenantIdentifiers: &tenant.TenantIdentifiers,
			}, creds)
			if err != nil {
				return err
			}
			meteringData.Tenants = append(meteringData.Tenants, &ttipb.MeteringData_TenantMeteringData{
				TenantIdentifiers: tenant.TenantIdentifiers,
				Totals:            totals,
			})
		}
		if len(res.Tenants) < tenantsPageSize {
			break
		}
		tenantsPage++
	}

	if len(meteringData.Tenants) == 0 {
		totals, err := reg.GetRegistryTotals(ctx, &ttipb.GetTenantRegistryTotalsRequest{}, creds)
		if err != nil {
			return err
		}
		meteringData.Tenants = append(meteringData.Tenants, &ttipb.MeteringData_TenantMeteringData{
			Totals: totals,
		})
	}

	if err = s.reporter.Report(ctx, &meteringData); err != nil {
		return err
	}

	if s.config.OnSuccess != nil {
		now := time.Now()
		s.mu.Lock()
		s.apply = func(license ttipb.License) ttipb.License {
			if s.config.OnSuccess.ExtendValidUntil != nil {
				license.ValidUntil = now.Add(*s.config.OnSuccess.ExtendValidUntil)
			}
			return license
		}
		s.mu.Unlock()
	}

	return nil
}

// Run the periodic metrics reporting.
func (s *meteringSetup) Run(ctx context.Context) error {
	interval := s.config.Interval
	if interval == 0 {
		interval = time.Hour
	}
	reportTicker := time.NewTicker(interval)
	defer reportTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-reportTicker.C:
			s.CollectAndReport(ctx)
		}
	}
}

func newMeteringSetup(ctx context.Context, config *ttipb.MeteringConfiguration, cluster Cluster) (*meteringSetup, error) {
	s := &meteringSetup{
		config:  config,
		cluster: cluster,
	}
	var err error
	switch reporterConfig := config.Metering.(type) {
	case *ttipb.MeteringConfiguration_AWS_:
		s.reporter, err = awsmetrics.New(reporterConfig.AWS)
		if err != nil {
			return nil, err
		}
	case *ttipb.MeteringConfiguration_Prometheus_:
		s.reporter, err = prometheusmetrics.New(reporterConfig.Prometheus, metrics.Registry)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported metering reporter config type: %T", config.Metering)
	}
	return s, nil
}

var globalMetering *meteringSetup

// SetupMetering sets up metering on cluster.
func SetupMetering(ctx context.Context, config *ttipb.MeteringConfiguration, cluster Cluster) error {
	if globalMetering != nil {
		return errors.New("only one metering configuration can be set up")
	}
	var err error
	globalMetering, err = newMeteringSetup(ctx, config, cluster)
	if err != nil {
		return err
	}
	if err := globalMetering.CollectAndReport(ctx); err != nil {
		return err
	}
	go globalMetering.Run(ctx)
	return nil
}
