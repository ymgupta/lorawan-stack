// Copyright © 2020 The Things Industries B.V.

package commands

import (
	"context"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/tenant"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
)

var (
	createAuthProviderCommand = &cobra.Command{
		Use:   "create-auth-provider",
		Short: "Create a federated authentication provider in the Identity Server database",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			logger.Info("Connecting to Identity Server database...")
			db, err := store.Open(ctx, config.IS.DatabaseURI)
			if err != nil {
				return err
			}
			defer db.Close()

			tenantID, err := cmd.Flags().GetString("tenant-id")
			if err != nil {
				return err
			}
			if tenantID == "" {
				tenantID = config.Tenancy.DefaultID
			}
			ctx = tenant.NewContext(ctx, ttipb.TenantIdentifiers{TenantID: tenantID})

			providerID, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			allowRegistrations, err := cmd.Flags().GetBool("allow-registrations")
			if err != nil {
				return err
			}

			fieldMask := &pbtypes.FieldMask{Paths: []string{
				"ids.provider_id",
				"name",
				"allow_registrations",
			}}
			provider := &ttipb.AuthenticationProvider{
				AuthenticationProviderIdentifiers: ttipb.AuthenticationProviderIdentifiers{ProviderID: providerID},
				Name:                              name,
				AllowRegistrations:                allowRegistrations,
			}

			if oidc, _ := cmd.Flags().GetBool("oidc"); oidc {
				clientID, err := cmd.Flags().GetString("oidc.client-id")
				if err != nil {
					return err
				}
				clientSecret, err := cmd.Flags().GetString("oidc.client-secret")
				if err != nil {
					return err
				}
				providerURL, err := cmd.Flags().GetString("oidc.provider-url")
				if err != nil {
					return err
				}

				provider.Configuration = &ttipb.AuthenticationProvider_Configuration{
					Provider: &ttipb.AuthenticationProvider_Configuration_OIDC{
						OIDC: &ttipb.AuthenticationProvider_OIDC{
							ClientID:     clientID,
							ClientSecret: clientSecret,
							ProviderURL:  providerURL,
						},
					},
				}
				fieldMask.Paths = append(fieldMask.Paths, "configuration.provider.oidc")
			}

			providerStore := store.GetAuthenticationProviderStore(db)

			var providerExists bool
			if _, err := providerStore.GetAuthenticationProvider(ctx, &provider.AuthenticationProviderIdentifiers, nil); err == nil {
				providerExists = true
			}

			if providerExists {
				logger.Info("Updating provider...")
				if _, err = providerStore.UpdateAuthenticationProvider(ctx, provider, fieldMask); err != nil {
					return err
				}
				logger.Info("Updated provider")
			} else {
				logger.Info("Creating provider...")
				if _, err = providerStore.CreateAuthenticationProvider(ctx, provider); err != nil {
					return err
				}
				logger.Info("Created provider")
			}

			return nil
		},
	}
)

func init() {
	createAuthProviderCommand.Flags().String("tenant-id", "", "Tenant ID")
	createAuthProviderCommand.Flags().Lookup("tenant-id").Hidden = true
	createAuthProviderCommand.Flags().String("id", "", "Provider ID")
	createAuthProviderCommand.Flags().String("name", "", "Provider Name")
	createAuthProviderCommand.Flags().Bool("allow-registrations", false, "Allow registrations on login")
	createAuthProviderCommand.Flags().Bool("oidc", false, "Use OIDC provider")
	createAuthProviderCommand.Flags().String("oidc.client-id", "", "OIDC Client ID")
	createAuthProviderCommand.Flags().String("oidc.client-secret", "", "OIDC Client Secret")
	createAuthProviderCommand.Flags().String("oidc.provider-url", "", "OIDC Provider URL")
	isDBCommand.AddCommand(createAuthProviderCommand)
}
