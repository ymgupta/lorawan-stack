// Copyright © 2019 The Things Industries B.V.

//+build tti

package commands

import (
	"context"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttipb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
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
			if tenantID == "" {
				tenantID = config.Tenancy.DefaultID
			}

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			tntFieldMask := &pbtypes.FieldMask{Paths: []string{
				"name",
			}}
			tnt := &ttipb.Tenant{
				TenantIdentifiers: ttipb.TenantIdentifiers{TenantID: tenantID},
				State:             ttnpb.STATE_APPROVED,
			}

			tntStore := store.GetTenantStore(db)

			var tntExists bool
			if _, err := tntStore.GetTenant(ctx, &tnt.TenantIdentifiers, tntFieldMask); err == nil {
				tntExists = true
			}

			tnt.Name = name

			if tntExists {
				logger.Info("Updating tenant...")
				if _, err = tntStore.UpdateTenant(ctx, tnt, tntFieldMask); err != nil {
					return err
				}
				logger.Info("Updated tenant")
			} else {
				logger.Info("Creating tenant...")
				if _, err = tntStore.CreateTenant(ctx, tnt); err != nil {
					return err
				}
				logger.Info("Created tenant")
			}

			return nil
		},
	}
)

func init() {
	createTenantCommand.Flags().String("id", "", "Tenant ID")
	createTenantCommand.Flags().String("name", "", "Name of the tenant")
	isDBCommand.AddCommand(createTenantCommand)
}
