package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v30/x/pegkeeper/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group pegkeeper queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdQueryPegState())
	cmd.AddCommand(CmdQueryAdjustmentHistory())

	return cmd
}

func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "shows the parameters of the module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

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

func CmdQueryPegState() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "peg-state",
		Short: "Query the current peg state",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.PegState(context.Background(), &types.QueryPegStateRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryAdjustmentHistory() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "adjustment-history [limit] [offset]",
		Short: "Query the supply adjustment history",
		Args:  cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryAdjustmentHistoryRequest{}
			if len(args) > 0 {
				limit, err := cmd.Flags().GetUint64("limit")
				if err != nil {
					return err
				}
				req.Limit = limit
			}
			if len(args) > 1 {
				offset, err := cmd.Flags().GetUint64("offset")
				if err != nil {
					return err
				}
				req.Offset = offset
			}

			res, err := queryClient.AdjustmentHistory(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().Uint64("limit", 10, "Maximum number of adjustments to return")
	cmd.Flags().Uint64("offset", 0, "Number of adjustments to skip")

	return cmd
}
