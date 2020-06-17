// Copyright © 2019 The Things Industries B.V.

package commands

import (
	"os"

	"github.com/gogo/protobuf/types"
	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/api"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/io"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttipb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

var (
	selectTenantFlags               = util.FieldMaskFlags(&ttipb.Tenant{})
	setTenantFlags                  = util.FieldFlags(&ttipb.Tenant{})
	stripeBillingProviderFlags      = util.FieldFlags(&ttipb.Billing_Stripe{}, "billing", "provider", "stripe")
	setInitialUserFlags             = util.FieldFlags(&ttnpb.User{}, "initial_user")
	selectTenantRegistryTotalsFlags = util.FieldMaskFlags(&ttipb.TenantRegistryTotals{})
)

func tenantIDFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.String("tenant-id", "", "")
	return flagSet
}

func tenantBillingFlags() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Bool("billing.provider.stripe", false, "use the Stripe billing provider")
	flagSet.AddFlagSet(stripeBillingProviderFlags)
	return flagSet
}

var errNoTenantID = errors.DefineInvalidArgument("no_tenant_id", "no tenant ID set")

func getTenantID(flagSet *pflag.FlagSet, args []string) *ttipb.TenantIdentifiers {
	var tenantID string
	if len(args) > 0 {
		if len(args) > 1 {
			logger.Warn("multiple IDs found in arguments, considering only the first")
		}
		tenantID = args[0]
	} else {
		tenantID, _ = flagSet.GetString("tenant-id")
	}
	if tenantID == "" {
		return nil
	}
	return &ttipb.TenantIdentifiers{TenantID: tenantID}
}

func getTenantAdminCreds(cmd *cobra.Command) grpc.CallOption {
	tenantAdminKey, _ := cmd.Flags().GetString("tenant-admin-key")
	return api.GetCredentials("TenantAdminKey", tenantAdminKey)
}

var (
	tenantsCommand = &cobra.Command{
		Use:               "tenants",
		Hidden:            true,
		Aliases:           []string{"tenant", "tnt", "t"},
		Short:             "Tenant commands",
		PersistentPreRunE: preRun(),
	}
	tenantsListCommand = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List tenants",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := util.SelectFieldMask(cmd.Flags(), selectTenantFlags)

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			limit, page, opt, getTotal := withPagination(cmd.Flags())
			res, err := ttipb.NewTenantRegistryClient(is).List(ctx, &ttipb.ListTenantsRequest{
				FieldMask: types.FieldMask{Paths: paths},
				Limit:     limit,
				Page:      page,
				Order:     getOrder(cmd.Flags()),
			}, opt, getTenantAdminCreds(cmd))
			if err != nil {
				return err
			}
			getTotal()

			return io.Write(os.Stdout, config.OutputFormat, res.Tenants)
		},
	}
	tenantsGetCommand = &cobra.Command{
		Use:     "get [tenant-id]",
		Aliases: []string{"info"},
		Short:   "Get a tenant",
		RunE: func(cmd *cobra.Command, args []string) error {
			refreshToken() // NOTE: ignore errors.
			optionalAuth()

			cliID := getTenantID(cmd.Flags(), args)
			if cliID == nil {
				return errNoTenantID
			}
			paths := util.SelectFieldMask(cmd.Flags(), selectTenantFlags)

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttipb.NewTenantRegistryClient(is).Get(ctx, &ttipb.GetTenantRequest{
				TenantIdentifiers: *cliID,
				FieldMask:         types.FieldMask{Paths: paths},
			}, getTenantAdminCreds(cmd))
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	tenantsCreateCommand = &cobra.Command{
		Use:     "create [tenant-id]",
		Aliases: []string{"add", "register"},
		Short:   "Create a tenant",
		RunE: asBulk(func(cmd *cobra.Command, args []string) (err error) {
			cliID := getTenantID(cmd.Flags(), args)
			var tenant ttipb.Tenant
			if inputDecoder != nil {
				_, err := inputDecoder.Decode(&tenant)
				if err != nil {
					return err
				}
			}
			if err := util.SetFields(&tenant, setTenantFlags); err != nil {
				return err
			}
			tenant.Attributes = mergeAttributes(tenant.Attributes, cmd.Flags())
			if cliID != nil && cliID.TenantID != "" {
				tenant.TenantID = cliID.TenantID
			}
			if tenant.TenantID == "" {
				return errNoTenantID
			}

			var initialUser *ttnpb.User
			if createInitialUser, _ := cmd.Flags().GetBool("initial_user"); createInitialUser {
				var user ttnpb.User
				if err := util.SetFields(&user, setInitialUserFlags, "initial_user"); err != nil {
					return err
				}
				if user.UserID == "" {
					return errNoUserID
				}

				if user.Password == "" {
					pw, err := gopass.GetPasswdPrompt("Please enter password:", true, os.Stdin, os.Stderr)
					if err != nil {
						return err
					}
					user.Password = string(pw)
					pw, err = gopass.GetPasswdPrompt("Please confirm password:", true, os.Stdin, os.Stderr)
					if err != nil {
						return err
					}
					if string(pw) != user.Password {
						return errPasswordMismatch
					}
				}

				initialUser = &user
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttipb.NewTenantRegistryClient(is).Create(ctx, &ttipb.CreateTenantRequest{
				Tenant:      tenant,
				InitialUser: initialUser,
			}, getTenantAdminCreds(cmd))
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		}),
	}
	tenantsUpdateCommand = &cobra.Command{
		Use:     "update [tenant-id]",
		Aliases: []string{"set"},
		Short:   "Update a tenant",
		RunE: func(cmd *cobra.Command, args []string) error {
			refreshToken() // NOTE: ignore errors.
			optionalAuth()

			cliID := getTenantID(cmd.Flags(), args)
			if cliID == nil {
				return errNoTenantID
			}
			paths := util.UpdateFieldMask(cmd.Flags(), setTenantFlags, attributesFlags(), stripeBillingProviderFlags)
			rawUnsetPaths, _ := cmd.Flags().GetStringSlice("unset")
			unsetPaths := util.NormalizePaths(rawUnsetPaths)

			if len(paths)+len(unsetPaths) == 0 {
				logger.Warn("No fields selected, won't update anything")
				return nil
			}
			if remainingPaths := ttnpb.ExcludeFields(paths, unsetPaths...); len(remainingPaths) != len(paths) {
				overlapPaths := ttnpb.ExcludeFields(paths, remainingPaths...)
				return errConflictingPaths.WithAttributes("field_mask_paths", overlapPaths)
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			tenant, err := ttipb.NewTenantRegistryClient(is).Get(ctx, &ttipb.GetTenantRequest{
				TenantIdentifiers: *cliID,
				FieldMask:         types.FieldMask{Paths: paths},
			}, getTenantAdminCreds(cmd))
			if err != nil && !errors.IsNotFound(err) {
				return err
			}
			if tenant == nil {
				tenant = &ttipb.Tenant{TenantIdentifiers: *cliID}
			}

			if err := util.SetFields(&tenant, setTenantFlags); err != nil {
				return err
			}
			tenant.Attributes = mergeAttributes(tenant.Attributes, cmd.Flags())
			tenant.TenantIdentifiers = *cliID

			if stripe, _ := cmd.Flags().GetBool("billing.provider.stripe"); stripe {
				if tenant.GetBilling().GetStripe() == nil {
					tenant.Billing = &ttipb.Billing{Provider: &ttipb.Billing_Stripe_{
						Stripe: &ttipb.Billing_Stripe{},
					}}
				}
				if err = util.SetFields(tenant.GetBilling().GetStripe(), stripeBillingProviderFlags, "billing", "provider", "stripe"); err != nil {
					return err
				}
			}

			if err := tenant.SetFields(tenant, ttnpb.ExcludeFields(paths, unsetPaths...)...); err != nil {
				return err
			}

			res, err := ttipb.NewTenantRegistryClient(is).Update(ctx, &ttipb.UpdateTenantRequest{
				Tenant:    *tenant,
				FieldMask: types.FieldMask{Paths: append(paths, unsetPaths...)},
			}, getTenantAdminCreds(cmd))
			if err != nil {
				return err
			}

			res.SetFields(tenant, "ids")
			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
	tenantsDeleteCommand = &cobra.Command{
		Use:   "delete [tenant-id]",
		Short: "Delete a tenant",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliID := getTenantID(cmd.Flags(), args)
			if cliID == nil {
				return errNoTenantID
			}

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			_, err = ttipb.NewTenantRegistryClient(is).Delete(ctx, cliID, getTenantAdminCreds(cmd))
			if err != nil {
				return err
			}

			return nil
		},
	}
	tenantsGetRegistryTotalsCommand = &cobra.Command{
		Use:     "get-registry-totals [tenant-id]",
		Aliases: []string{"totals"},
		Short:   "Get registry totals of a tenant",
		RunE: func(cmd *cobra.Command, args []string) error {
			tntID := getTenantID(cmd.Flags(), args)
			paths := util.SelectFieldMask(cmd.Flags(), selectTenantRegistryTotalsFlags)

			is, err := api.Dial(ctx, config.IdentityServerGRPCAddress)
			if err != nil {
				return err
			}
			res, err := ttipb.NewTenantRegistryClient(is).GetRegistryTotals(ctx, &ttipb.GetTenantRegistryTotalsRequest{
				TenantIdentifiers: tntID,
				FieldMask:         types.FieldMask{Paths: paths},
			}, getTenantAdminCreds(cmd))
			if err != nil {
				return err
			}

			return io.Write(os.Stdout, config.OutputFormat, res)
		},
	}
)

func init() {
	tenantsListCommand.Flags().AddFlagSet(selectTenantFlags)
	tenantsListCommand.Flags().AddFlagSet(paginationFlags())
	tenantsListCommand.Flags().AddFlagSet(orderFlags())
	tenantsCommand.AddCommand(tenantsListCommand)
	tenantsGetCommand.Flags().AddFlagSet(tenantIDFlags())
	tenantsGetCommand.Flags().AddFlagSet(selectTenantFlags)
	tenantsCommand.AddCommand(tenantsGetCommand)
	tenantsCreateCommand.Flags().AddFlagSet(tenantIDFlags())
	tenantsCreateCommand.Flags().AddFlagSet(setTenantFlags)
	tenantsCreateCommand.Flags().Bool("initial_user", false, "create initial tenant user")
	tenantsCreateCommand.Flags().AddFlagSet(setInitialUserFlags)
	tenantsCreateCommand.Flags().Lookup("initial_user.state").DefValue = ttnpb.STATE_APPROVED.String()
	tenantsCreateCommand.Flags().AddFlagSet(attributesFlags())
	tenantsCommand.AddCommand(tenantsCreateCommand)
	tenantsUpdateCommand.Flags().AddFlagSet(tenantIDFlags())
	tenantsUpdateCommand.Flags().AddFlagSet(setTenantFlags)
	tenantsUpdateCommand.Flags().AddFlagSet(tenantBillingFlags())
	tenantsUpdateCommand.Flags().AddFlagSet(attributesFlags())
	tenantsUpdateCommand.Flags().AddFlagSet(util.UnsetFlagSet())
	tenantsCommand.AddCommand(tenantsUpdateCommand)
	tenantsDeleteCommand.Flags().AddFlagSet(tenantIDFlags())
	tenantsCommand.AddCommand(tenantsDeleteCommand)
	tenantsGetRegistryTotalsCommand.Flags().AddFlagSet(tenantIDFlags())
	tenantsGetRegistryTotalsCommand.Flags().AddFlagSet(selectTenantRegistryTotalsFlags)
	tenantsCommand.AddCommand(tenantsGetRegistryTotalsCommand)
	tenantsCommand.PersistentFlags().String("tenant-admin-key", "", "Tenant Admin Key")
	Root.AddCommand(tenantsCommand)
}
