// Copyright © 2019 The Things Industries B.V.

package component

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

// WithLicense returns an option to configure the Component with a License.
func WithLicense(l ttipb.License) Option {
	return func(c *Component) {
		c.ctx = license.NewContextWithLicense(c.ctx, l)
		c.AddContextFiller(func(ctx context.Context) context.Context {
			return license.NewContextWithLicense(ctx, l)
		})
	}
}
