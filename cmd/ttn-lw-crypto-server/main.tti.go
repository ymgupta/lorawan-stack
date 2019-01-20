// Copyright © 2019 The Things Industries B.V.

// ttn-lw-crypto-server is the binary that hosts a crypto service of The Things Network Stack for LoRaWAN.
package main

import (
	"os"

	"go.thethings.network/lorawan-stack/cmd/internal/errors"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-crypto-server/commands"
)

func main() {
	if err := commands.Root.Execute(); err != nil {
		errors.PrintStack(os.Stderr, err)
		os.Exit(-1)
	}
}
