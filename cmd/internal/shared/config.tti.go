// Copyright © 2019 The Things Industries B.V.

package shared

import (
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
)

// DefaultTenancyConfig is the default tenancy configuration.
var DefaultTenancyConfig = tenant.Config{
	DefaultID: "default",
	CacheTTL:  time.Minute,
}

// DefaultDeviceClaimingPublicURL is the default public URL where the device claiming is served.
var DefaultDeviceClaimingPublicURL = DefaultPublicURL + "/claim"
