// Copyright © 2019 The Things Industries B.V.

package component_test

import "go.thethings.network/lorawan-stack/pkg/tenant"

func init() {
	tenant.AllowEmptyTenantID()
}
