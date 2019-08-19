// Copyright © 2019 The Things Industries B.V.

package license

import (
	"context"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

type licenseContextKeyType struct{}

var licenseContextKey licenseContextKeyType

type licenseContext struct {
	mu      sync.RWMutex
	license *ttipb.License
}

var now = time.Now()

var defaultLicense = ttipb.License{
	LicenseIdentifiers:      ttipb.LicenseIdentifiers{LicenseID: "development"},
	CreatedAt:               now,
	ValidFrom:               now,
	ValidUntil:              now.Add(time.Hour),
	ComponentAddressRegexps: []string{"localhost"},
	DevAddrPrefixes: []types.DevAddrPrefix{
		{DevAddr: types.DevAddr{0, 0, 0, 0}, Length: 7},
		{DevAddr: types.DevAddr{0, 0, 0, 2}, Length: 7},
	},
}

// NewContextWithLicense returns a context derived from parent that contains license.
func NewContextWithLicense(parent context.Context, license ttipb.License) context.Context {
	return context.WithValue(parent, licenseContextKey, &licenseContext{license: &license})
}

// FromContext returns the License from the context if present. Otherwise it returns default License.
func FromContext(ctx context.Context) ttipb.License {
	if lc, ok := ctx.Value(licenseContextKey).(*licenseContext); ok {
		lc.mu.RLock()
		license := *lc.license
		lc.mu.RUnlock()
		return license
	}
	return defaultLicense
}

// Mutate mutates the license in the context.
func Mutate(ctx context.Context, update func(license ttipb.License) ttipb.License) {
	lc, ok := ctx.Value(licenseContextKey).(*licenseContext)
	if !ok {
		panic("no license in context")
	}
	lc.mu.Lock()
	license := *lc.license
	license = update(license)
	lc.license = &license
	lc.mu.Unlock()
}
