// Copyright © 2020 The Things Industries B.V.

package networkserver

import (
	"go.thethings.network/lorawan-stack/pkg/errors"
)

var (
	errNoTenant = errors.DefineNotFound("no_tenant", "no tenant present in context")
)
