// Copyright © 2019 The Things Industries B.V.

package cryptoserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/cryptoserver/provisioners/microchip"
)

// Config represents the CryptoServer configuration.
type Config struct {
	Microchip microchip.Config `name:"microchip"`
}

// NewProvisioners returns a new Provisioners from the configuration.
func (c *Config) NewProvisioners(ctx context.Context) (Provisioners, error) {
	provisioners := provisioners(make(map[string]Provisioner))
	if c.Microchip.Enable {
		service, err := microchip.New(ctx, &c.Microchip)
		if err != nil {
			return nil, err
		}
		provisioners["microchip"] = Provisioner{
			Network:        service,
			Application:    service,
			ExposeRootKeys: c.Microchip.ExposeRootKeys,
		}
	}
	return provisioners, nil
}
