// Copyright © 2019 The Things Industries B.V.

package blacklist

func init() {
	builtin.Add(
		"tbs",
		"tenantbillingserver",
		"tti-lw-cli",
		"tti-lw-stack",
	)
}
