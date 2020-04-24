// Copyright © 2020 The Things Industries B.V.

package packetbrokeragent_test

import (
	. "go.thethings.network/lorawan-stack/pkg/packetbrokeragent"
)

func init() {
	testOptions = append(testOptions, WithTenancyContextFiller())
}
