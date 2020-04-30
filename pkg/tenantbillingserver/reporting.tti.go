// Copyright © 2020 The Things Industries B.V.

package tenantbillingserver

import (
	"context"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
)

var errStripeDisabled = errors.DefineFailedPrecondition("stripe_disabled", "the Stripe billing backend is disabled")

func (tbs *TenantBillingServer) contactBackend(ctx context.Context, tnt *ttipb.Tenant, totals *ttipb.TenantRegistryTotals) error {
	if tnt.Billing == nil {
		return nil
	}
	switch tnt.Billing.Provider.(type) {
	case *ttipb.Billing_Stripe_:
		if tbs.backends.stripe == nil {
			return errStripeDisabled.New()
		}
		return tbs.backends.stripe.Report(ctx, tnt, totals)
	}
	return nil
}

func (tbs *TenantBillingServer) collectAndReport(ctx context.Context) error {
	cc, err := tbs.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		return err
	}

	var (
		tenantsPageSize = 1000
		tenantsPage     = 1

		logger   = log.FromContext(ctx)
		registry = ttipb.NewTenantRegistryClient(cc)
		creds    = grpc.PerRPCCredentials(rpcmetadata.MD{
			AuthType:  tenantAdminAuthType,
			AuthValue: tbs.config.TenantAdminKey,
		})
	)
	for {
		res, err := registry.List(ctx, &ttipb.ListTenantsRequest{
			FieldMask: pbtypes.FieldMask{Paths: []string{"ids", "billing", "state"}},
			Limit:     uint32(tenantsPageSize),
			Page:      uint32(tenantsPage),
		}, creds)
		if err != nil {
			return err
		}
		for _, tenant := range res.Tenants {
			totals, err := registry.GetRegistryTotals(ctx, &ttipb.GetTenantRegistryTotalsRequest{
				TenantIdentifiers: &tenant.TenantIdentifiers,
			}, creds)
			if err != nil {
				return err
			}
			err = tbs.contactBackend(ctx, tenant, totals)
			if err != nil {
				logger.WithError(err).Error("Failed to report metrics to backend")
				continue
			}
		}
		if len(res.Tenants) < tenantsPageSize {
			break
		}
		tenantsPage++
	}
	return nil
}

func (tbs *TenantBillingServer) run(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger.WithField("pull_interval", tbs.config.PullInterval).Debug("Periodic metering data worker started")

	reportTicker := time.NewTicker(tbs.config.PullInterval)
	defer reportTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-reportTicker.C:
			if err := tbs.collectAndReport(ctx); err != nil {
				logger.WithError(err).Error("Failed to collect and report the tenant metering data")
			}
		}
	}
}
