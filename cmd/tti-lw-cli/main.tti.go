// Copyright © 2019 The Things Industries B.V.

// tti-lw-cli is the binary for the Command-line interface of The Things Industries LoRaWAN Stack.
package main

import (
	"fmt"
	"os"

	cli_errors "go.thethings.network/lorawan-stack/v3/cmd/internal/errors"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/commands"
	"go.thethings.network/lorawan-stack/v3/config/tags"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

func main() {
	if !tags.TTI {
		fmt.Fprintln(os.Stderr, "Can't run tti-lw-cli without the tti build tag")
		os.Exit(2)
	}
	commands.Root.Short = "The Things Industries Command-line Interface"
	if err := commands.Root.Execute(); err != nil {
		if errors.IsCanceled(err) {
			os.Exit(130)
		}
		cli_errors.PrintStack(os.Stderr, err)
		os.Exit(-1)
	}
}
