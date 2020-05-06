// Copyright © 2020 The Things Industries B.V.

package packetbrokeragent_test

import (
	. "go.thethings.network/lorawan-stack/v3/pkg/packetbrokeragent"
)

func init() {
	testOptions = append(testOptions, WithTenancyContextFiller())
}
