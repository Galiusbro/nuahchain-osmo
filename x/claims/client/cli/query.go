package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/osmosis-labs/osmosis/v30/x/claims/types"
)

// GetQueryCmd returns the root query command for the claims module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "claims",
		Short:                      "Query commands for the claims module",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(
		newQueryClaimCmd(),
		newQueryClaimsCmd(),
		newQueryParamsCmd(),
	)

	return cmd
}

func newQueryClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim [claim-id]",
		Short: "Fetch a claim by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			claimID, err := parseUint64(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Claim(cmd.Context(), &types.QueryClaimRequest{ClaimId: claimID})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func newQueryClaimsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claims",
		Short: "List claims with optional filters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			policyID, _ := cmd.Flags().GetUint64(flagPolicyID)
			claimant, _ := cmd.Flags().GetString(flagClaimant)
			statusStr, _ := cmd.Flags().GetString(flagStatus)

			filter := &types.QueryClaimsRequest{
				PolicyId: policyID,
				Claimant: claimant,
			}

			if statusStr != "" {
				status, err := parseClaimStatus(statusStr)
				if err != nil {
					return err
				}
				filter.Status = status
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			filter.Pagination = pageReq

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Claims(cmd.Context(), filter)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Uint64(flagPolicyID, 0, "Filter by policy ID")
	cmd.Flags().String(flagClaimant, "", "Filter by claimant address")
	cmd.Flags().String(flagStatus, "", "Filter by status (pending, approved, rejected, paid)")
	flags.AddPaginationFlagsToCmd(cmd, "claims")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func newQueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Show the current claims module parameters",
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

// const (
// 	flagPolicyID = "policy-id"
// 	flagClaimant = "claimant"
// 	flagStatus   = "status"
// )
