// Copyright © 2019 The Things Industries B.V.

package commands

import (
	"context"

	pkglicense "go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

var license *ttipb.License

func initializeLicense(ctx context.Context) (context.Context, error) {
	var err error
	license, err = pkglicense.Read(config.License)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx)
	if license == nil {
		logger.Warn("No license configured, running in unlicensed mode")
		return ctx, nil
	}
	logger.WithFields(license).Info("Valid license")
	ctx = pkglicense.NewContextWithLicense(ctx, *license)
	return ctx, err
}
