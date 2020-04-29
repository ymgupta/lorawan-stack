// Copyright © 2020 The Things Industries B.V.

package migrations

func init() {
	All = append(All,
		TenantStripeAttributeBilling{},
	)
}
