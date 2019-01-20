// Copyright © 2019 The Things Industries B.V.

// Package commands implements the commands for the ttn-lw-crypto-server binary.
package commands

import (
	"os"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/cmd/internal/shared/version"
	conf "go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/log"
)

var (
	logger *log.Logger
	name   = "ttn-lw-crypto-server"
	mgr    = conf.InitializeWithDefaults(name, "ttn_lw", DefaultConfig)
	config = new(Config)

	// Root command is the entrypoint of the program
	Root = &cobra.Command{
		Use:           name,
		SilenceErrors: true,
		SilenceUsage:  true,
		Short:         "The Things Network Crypto Server for LoRaWAN",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// read in config from file
			err := mgr.ReadInConfig()
			if err != nil {
				return err
			}

			// unmarshal config
			if err = mgr.Unmarshal(config); err != nil {
				return err
			}

			// create logger
			logger, err = log.NewLogger(
				log.WithLevel(config.Log.Level),
				log.WithHandler(log.NewCLI(os.Stdout)),
			)
			if sentry, err := shared.SentryMiddleware(config.ServiceBase); err == nil && sentry != nil {
				logger.Use(sentry)
			}

			// initialize shared packages
			if err := shared.Initialize(config.ServiceBase); err != nil {
				return err
			}

			return err
		},
	}
)

func init() {
	Root.PersistentFlags().AddFlagSet(mgr.Flags())
	Root.AddCommand(version.Print(name))
}
