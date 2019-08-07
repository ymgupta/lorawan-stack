// Copyright © 2019 The Things Industries B.V.

// tti-lw-stack is the binary that runs the entire The Things Industries LoRaWAN Stack.
package main

import (
	"fmt"
	"os"

	"go.thethings.network/lorawan-stack/cmd/internal/errors"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-stack/commands"
	"go.thethings.network/lorawan-stack/config/tags"
)

func main() {
	if !tags.TTI {
		fmt.Fprintln(os.Stderr, "Can't run tti-lw-stack without the tti build tag")
		os.Exit(2)
	}
	commands.Root.Short = "The Things Industries LoRaWAN Stack"
	if err := commands.Root.Execute(); err != nil {
		errors.PrintStack(os.Stderr, err)
		os.Exit(-1)
	}
}
