// Copyright © 2019 The Things Industries B.V.

//+build tti

package commands

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
)

var (
	createTenantCommand = &cobra.Command{
		Use:    "create-tenant",
		Short:  "Create a tenant in the Identity Server database",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(log.NewContext(context.Background(), logger), 10*time.Second)
			defer cancel()

			logger.Info("Connecting to Identity Server database...")
			db, err := store.Open(ctx, config.IS.DatabaseURI)
			if err != nil {
				return err
			}
			defer db.Close()

			tenantID, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			logger.Info("Creating tenant...")
			tenantStore := store.GetTenantStore(db)
			_, err = tenantStore.CreateTenant(ctx, &ttipb.Tenant{
				TenantIdentifiers: ttipb.TenantIdentifiers{TenantID: tenantID},
				Name:              name,
			})
			if err != nil {
				return err
			}

			logger.Info("Created tenant")
			return nil
		},
	}
)

func init() {
	createTenantCommand.Flags().String("id", "", "Tenant ID")
	createTenantCommand.Flags().String("name", "", "Name of the tenant")
	isDBCommand.AddCommand(createTenantCommand)
}
