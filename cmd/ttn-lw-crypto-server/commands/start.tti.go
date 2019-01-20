// Copyright © 2019 The Things Industries B.V.

package commands

import (
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/cryptoserver"
)

var (
	startCommand = &cobra.Command{
		Use:   "start",
		Short: "Start the Crypto Server",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := component.New(logger, &component.Config{ServiceBase: config.ServiceBase})
			if err != nil {
				return shared.ErrInitializeBaseComponent.WithCause(err)
			}

			cs, err := cryptoserver.New(c, &config.CS)
			if err != nil {
				return shared.ErrInitializeCryptoServer.WithCause(err)
			}
			_ = cs

			logger.Info("Starting Crypto Server...")
			return c.Run()
		},
	}
)

func init() {
	Root.AddCommand(startCommand)
}
