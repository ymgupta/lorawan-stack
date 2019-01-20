// Copyright © 2019 The Things Industries B.V.

package cryptoserver

import (
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoservices"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

// Config represents the CryptoServer configuration.
type Config struct {
	NetworkService     string `name:"network-service" description:"Network crypto service to host"`
	ApplicationService string `name:"application-service" description:"Application crypto service to host"`
}

var errServiceNotConfigured = errors.DefineFailedPrecondition("service_not_configured", "service `{id}` not configured")

// Network returns the service for network-layer crypto operations.
func (c *Config) Network() (cryptoservices.Network, error) {
	if c.NetworkService == "" {
		return nil, nil
	}
	return nil, nil
}

// Application returns the service for application-layer crypto operations.
func (c *Config) Application() (cryptoservices.Application, error) {
	if c.ApplicationService == "" {
		return nil, nil
	}
	return nil, nil
}
