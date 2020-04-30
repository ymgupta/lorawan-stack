// Copyright © 2019 The Things Industries B.V.

package component

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/license"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
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

type meteredComponent struct {
	*Component
}

func (c meteredComponent) Auth() grpc.CallOption {
	return c.cluster.Auth()
}

func (c meteredComponent) GetPeerConn(ctx context.Context, role ttnpb.ClusterRole) (*grpc.ClientConn, error) {
	return c.cluster.GetPeerConn(ctx, role, nil)
}

func (c *Component) startMetering() error {
	l := license.FromContext(c.Context())
	if l.Metering != nil {
		return license.SetupMetering(c.Context(), l.Metering, meteredComponent{c})
	}
	return nil
}
