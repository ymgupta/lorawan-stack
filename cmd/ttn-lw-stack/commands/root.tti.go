// Copyright © 2019 The Things Industries B.V.

package commands

import (
	"context"

	pkglicense "go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

var license *ttipb.License

func initializeLicense(ctx context.Context) (context.Context, error) {
	var err error
	license, err = pkglicense.Read(config.License)
	if err == nil && license != nil {
		ctx = pkglicense.NewContextWithLicense(ctx, *license)
	}
	return ctx, err
}
