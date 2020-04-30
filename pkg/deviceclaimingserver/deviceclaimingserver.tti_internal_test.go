// Copyright © 2019 The Things Industries B.V.

package deviceclaimingserver

import (
	"strings"
	"time"
)

func init() {
	customSameHost = strings.EqualFold
	dialTimeout = 1 * time.Second
}
