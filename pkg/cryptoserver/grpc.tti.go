// Copyright © 2019 The Things Industries B.V.

package cryptoserver

import "go.thethings.network/lorawan-stack/pkg/errors"

var (
	errServiceNotSupported = errors.DefineFailedPrecondition("service_not_supported", "crypto service not supported")
	errRootKeysNotExposed  = errors.DefinePermissionDenied("root_keys_not_exposed", "root keys are not being exposed")
)
