package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/osmosis-labs/osmosis/v30/x/policy/types"
)

// GetQueryCmd returns the root query command for the policy module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "policy",
		Short:                      "Query commands for the policy module",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(
		newQueryPolicyCmd(),
		newQueryPoliciesCmd(),
		newQueryParamsCmd(),
	)

	return cmd
}

func newQueryPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy [policy-id]",
		Short: "Fetch a policy by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			id, err := parseUint64(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Policy(cmd.Context(), &types.QueryPolicyRequest{PolicyId: id})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func newQueryPoliciesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policies",
		Short: "List policies with optional filters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			owner, _ := cmd.Flags().GetString(flagOwner)
			policyType, _ := cmd.Flags().GetString(flagPolicyType)
			statusStr, _ := cmd.Flags().GetString(flagStatus)

			filter := &types.PolicyFilter{}
			if owner != "" {
				filter.Owner = owner
			}
			if policyType != "" {
				filter.PolicyType = policyType
			}
			if statusStr != "" {
				status, err := parseStatus(statusStr)
				if err != nil {
					return err
				}
				filter.Status = status
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Policies(cmd.Context(), &types.QueryPoliciesRequest{
				Filter:     filter,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String(flagOwner, "", "Filter by owner address")
	cmd.Flags().String(flagPolicyType, "", "Filter by policy type")
	cmd.Flags().String(flagStatus, "", "Filter by status (active, expired, claimed, cancelled)")
	flags.AddPaginationFlagsToCmd(cmd, "policies")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func newQueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Show the current policy module parameters",
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
