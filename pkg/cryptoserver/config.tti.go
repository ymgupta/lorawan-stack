// Copyright © 2019 The Things Industries B.V.

package cryptoserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoservices"
	"go.thethings.network/lorawan-stack/pkg/cryptoserver/providers/microchip"
)

// Config represents the CryptoServer configuration.
type Config struct {
	NetworkService     string `name:"network-service" description:"Network crypto service to host (microchip)"`
	ApplicationService string `name:"application-service" description:"Application crypto service to host (microchip)"`

	ExposeRootKeys bool `name:"expose-root-keys" description:"Expose LoRaWAN root keys"`

	Microchip microchip.Config `name:"microchip"`
}

// NewNetwork returns a new service for network-layer crypto operations.
func (c *Config) NewNetwork(ctx context.Context) (cryptoservices.Network, error) {
	switch c.NetworkService {
	case "microchip":
		return microchip.New(ctx, &c.Microchip)
	}
	return nil, nil
}

// NewApplication returns a new service for application-layer crypto operations.
func (c *Config) NewApplication(ctx context.Context) (cryptoservices.Application, error) {
	switch c.ApplicationService {
	case "microchip":
		return microchip.New(ctx, &c.Microchip)
	}
	return nil, nil
}
