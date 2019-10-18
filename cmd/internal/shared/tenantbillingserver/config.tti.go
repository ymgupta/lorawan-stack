// Copyright © 2019 The Things Industries B.V.

package tenantbillingserver

import (
	"go.thethings.network/lorawan-stack/pkg/tenantbillingserver"
	"go.thethings.network/lorawan-stack/pkg/tenantbillingserver/stripe"
)

// DefaultTenantBillingServerConfig is the default configuration for the Tenant Billing Server.
var DefaultTenantBillingServerConfig = tenantbillingserver.Config{
	Stripe: stripe.Config{
		Enabled: false,
	},
}
