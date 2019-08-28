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

func (c meteredComponent) GetPeerConn(ctx context.Context, role ttnpb.ClusterRole) (*grpc.ClientConn, error) {
	peer, err := c.GetPeer(ctx, role, nil)
	if err != nil {
		return nil, err
	}
	conn, err := peer.Conn()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (c *Component) startMetering() error {
	l := license.FromContext(c.Context())
	if l.Metering != nil {
		return license.SetupMetering(c.Context(), l.Metering, meteredComponent{c})
	}
	return nil
}
