// Copyright © 2019 The Things Industries B.V.

// +build tti

package test

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/tenant"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

func init() {
	DefaultContext = tenant.NewContext(context.Background(), ttipb.TenantIdentifiers{TenantID: "foo-tenant"})
}
