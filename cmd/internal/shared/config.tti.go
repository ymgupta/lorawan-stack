// Copyright © 2019 The Things Industries B.V.

//+build tti

package shared

import (
	"go.thethings.network/lorawan-stack/pkg/config"
)

// DefaultTenancyConfig is the default tenancy configuration.
var DefaultTenancyConfig = config.Tenancy{
	DefaultID: "default",
}
