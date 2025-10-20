package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/osmosis-labs/osmosis/v30/x/assets/types"
)

// GetQueryCmd returns the root query command for the assets module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "assets",
		Short:                      "Query commands for the assets module",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(
		newAssetCmd(),
		newAssetsCmd(),
	)

	return cmd
}

func newAssetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset [symbol]",
		Short: "Fetch an asset by symbol",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			symbol := strings.TrimSpace(args[0])
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Asset(cmd.Context(), &types.QueryAssetRequest{
				Symbol: symbol,
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

func newAssetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assets",
		Short: "List registered assets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Assets(cmd.Context(), &types.QueryAssetsRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "assets")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
