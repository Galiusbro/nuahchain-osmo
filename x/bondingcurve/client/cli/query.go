package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
)

// GetQueryCmd returns the root query command for the bondingcurve module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Query commands for the bonding curve module",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(
		newQueryParamsCmd(),
		newQueryGlobalPauseCmd(),
		newQueryTokenPauseCmd(),
		newQueryFreezeCmd(),
		newQueryPendingParamsCmd(),
		newQueryEmergencyActionsCmd(),
		newQueryEmergencyConfigCmd(),
	)

	return cmd
}

func newQueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Show module parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func newQueryGlobalPauseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "global-pause",
		Short: "Show the global pause state",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.GlobalPause(cmd.Context(), &types.QueryGlobalPauseRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func newQueryTokenPauseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token-pause [denom]",
		Short: "Show pause status for a specific token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.TokenPause(cmd.Context(), &types.QueryTokenPauseRequest{Denom: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func newQueryFreezeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "freeze [target-type] [target]",
		Short: "Show freeze status for an address or token",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			targetTypeStr := strings.ToLower(args[0])
			var targetType types.FreezeTargetType
			switch targetTypeStr {
			case "address", "addr":
				targetType = types.FreezeTargetType_FREEZE_TARGET_TYPE_ADDRESS
			case "token", "denom":
				targetType = types.FreezeTargetType_FREEZE_TARGET_TYPE_TOKEN
			default:
				return fmt.Errorf("unknown target type %q", targetTypeStr)
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Freeze(cmd.Context(), &types.QueryFreezeRequest{TargetType: targetType, Target: args[1]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func newQueryPendingParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-params",
		Short: "Show pending parameter changes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.PendingParams(cmd.Context(), &types.QueryPendingParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func newQueryEmergencyActionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "emergency-actions",
		Short: "List recent emergency actions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			limitStr, _ := cmd.Flags().GetString("limit")
			var limit uint64
			if limitStr != "" {
				parsed, err := strconv.ParseUint(limitStr, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid limit: %w", err)
				}
				limit = parsed
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.EmergencyActions(cmd.Context(), &types.QueryEmergencyActionsRequest{Limit: limit})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("limit", "", "Optional maximum number of records to return")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func newQueryEmergencyConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "emergency-config",
		Short: "Show emergency signer configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.EmergencyConfig(cmd.Context(), &types.QueryEmergencyConfigRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
