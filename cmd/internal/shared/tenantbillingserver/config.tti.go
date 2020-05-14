// Copyright © 2019 The Things Industries B.V.

package tenantbillingserver

import (
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/tenantbillingserver"
)

// DefaultTenantBillingServerConfig is the default configuration for the Tenant Billing Server.
var DefaultTenantBillingServerConfig = tenantbillingserver.Config{
	PullInterval: 1 * time.Hour,
}
