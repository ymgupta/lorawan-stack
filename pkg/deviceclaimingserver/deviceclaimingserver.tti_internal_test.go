// Copyright © 2019 The Things Industries B.V.

package deviceclaimingserver

import (
	"strings"
)

func init() {
	customSameHost = strings.EqualFold
}
