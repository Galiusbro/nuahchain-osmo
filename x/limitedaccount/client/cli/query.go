// Package cli provides the command line interface for the limitedaccount module.
package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	limitedaccounttypes "github.com/osmosis-labs/osmosis/v30/x/limitedaccount/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        limitedaccounttypes.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", limitedaccounttypes.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryLimitedAccount())
	cmd.AddCommand(CmdQueryAllLimitedAccounts())

	return cmd
}

// CmdQueryLimitedAccount returns a CLI command handler for querying a limited account
func CmdQueryLimitedAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "limited-account [address]",
		Short: "Query a limited account by address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Validate address
			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("invalid address: %w", err)
			}

			queryClient := limitedaccounttypes.NewQueryClient(clientCtx)

			res, err := queryClient.LimitedAccount(context.Background(), &limitedaccounttypes.QueryLimitedAccountRequest{
				Address: addr.String(),
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdQueryAllLimitedAccounts returns a CLI command handler for querying all limited accounts
func CmdQueryAllLimitedAccounts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all-limited-accounts",
		Short: "Query all limited accounts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := limitedaccounttypes.NewQueryClient(clientCtx)

			res, err := queryClient.AllLimitedAccounts(context.Background(), &limitedaccounttypes.QueryAllLimitedAccountsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
