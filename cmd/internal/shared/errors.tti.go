// Copyright © 2019 The Things Industries B.V.

package shared

import "go.thethings.network/lorawan-stack/pkg/errors"

// Errors returned by component initialization.
var (
	ErrInitializeCryptoServer = errors.Define("initialize_crypto_server", "could not initialize Crypto Server")
)
