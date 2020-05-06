// Copyright © 2019 The Things Industries B.V.

package tenantbillingserver

import (
	"go.thethings.network/lorawan-stack/v3/pkg/tenantbillingserver"
)

// DefaultTenantBillingServerConfig is the default configuration for the Tenant Billing Server.
var DefaultTenantBillingServerConfig = tenantbillingserver.Config{
	ReporterAddressRegexps: []string{
		"localhost",
		"pipe",
	},
}
