package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/osmosis-labs/osmosis/v30/x/exchange/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group exchange queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdQueryExchangeRate())
	cmd.AddCommand(CmdQueryExchangeRates())
	cmd.AddCommand(CmdQueryDailyLimit())

	return cmd
}

// CmdQueryParams implements the params query command.
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the current exchange parameters",
		Long: fmt.Sprintf(`Query the current exchange parameters.

Example:
$ %s query %s params
`,
			version.AppName, types.ModuleName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdQueryExchangeRate implements the exchange rate query command.
func CmdQueryExchangeRate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exchange-rate [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Query exchange rate for a specific token",
		Long: fmt.Sprintf(`Query the current exchange rate for a specific token denomination.

Example:
$ %s query %s exchange-rate uosmo
`,
			version.AppName, types.ModuleName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.ExchangeRate(context.Background(), &types.QueryExchangeRateRequest{
				Denom: args[0],
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

// CmdQueryExchangeRates implements the exchange rates query command.
func CmdQueryExchangeRates() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exchange-rates",
		Args:  cobra.NoArgs,
		Short: "Query all exchange rates",
		Long: fmt.Sprintf(`Query all current exchange rates with pagination support.

Example:
$ %s query %s exchange-rates
$ %s query %s exchange-rates --page=2 --limit=50
`,
			version.AppName, types.ModuleName, version.AppName, types.ModuleName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.ExchangeRates(context.Background(), &types.QueryExchangeRatesRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)

	return cmd
}

// CmdQueryDailyLimit implements the daily limit query command.
func CmdQueryDailyLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daily-limit [address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query daily exchange limit for an address",
		Long: fmt.Sprintf(`Query the current daily exchange limit and usage for a specific address.

Example:
$ %s query %s daily-limit cosmos1abc123...
`,
			version.AppName, types.ModuleName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.DailyLimit(context.Background(), &types.QueryDailyLimitRequest{
				Address: args[0],
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