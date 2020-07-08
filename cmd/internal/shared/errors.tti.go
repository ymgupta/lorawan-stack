// Copyright © 2019 The Things Industries B.V.

package shared

import "go.thethings.network/lorawan-stack/v3/pkg/errors"

// Errors returned by component initialization.
var (
	ErrInitializeDeviceClaimingServer = errors.Define("initialize_device_claiming_server", "could not initialize Device Claiming Server")
	ErrInitializeCryptoServer         = errors.Define("initialize_crypto_server", "could not initialize Crypto Server")
	ErrInitializeTenantBillingServer  = errors.Define("initialize_tenant_billing_server", "could not initialize Tenant Billing Server")
	ErrInitializeEventServer          = errors.Define("initialize_event_server", "could not initialize Event Server")
)
