// Copyright © 2019 The Things Industries B.V.

package cryptoserver

import (
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoservices"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

// Provisioner contains network and application layer cryptographic services.
type Provisioner struct {
	Network        cryptoservices.Network
	Application    cryptoservices.Application
	ExposeRootKeys bool
}

// Provisioners gets a provisioner by provisioner ID.
type Provisioners interface {
	Get(id string) (Provisioner, error)
}

type provisioners map[string]Provisioner

var errProvisionerNotFound = errors.DefineNotFound("provisioner_not_found", "provisioner `{id}` not found")

func (p provisioners) Get(id string) (Provisioner, error) {
	res, ok := p[id]
	if !ok {
		return Provisioner{}, errProvisionerNotFound.WithAttributes("id", id)
	}
	return res, nil
}
