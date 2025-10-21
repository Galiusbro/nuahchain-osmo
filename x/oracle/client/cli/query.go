package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/osmosis-labs/osmosis/v30/x/oracle/types"
)

// GetQueryCmd returns the root query command for the oracle module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Query commands for the oracle module",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(NewPriceCmd())

	return cmd
}

// NewPriceCmd returns a query command for price lookups.
func NewPriceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price [symbol]",
		Short: "Query the price for a symbol",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			symbol := strings.TrimSpace(args[0])
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Price(cmd.Context(), &types.QueryPriceRequest{Symbol: symbol})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
