package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/osmosis-labs/osmosis/v30/x/premium/types"
)

// GetQueryCmd returns the root query command for the premium module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "premium",
		Short:                      "Query commands for the premium module",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(
		newQueryPlanCmd(),
		newQueryPlansCmd(),
		newQueryPaymentsCmd(),
		newQueryParamsCmd(),
	)

	return cmd
}

func newQueryPlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan [plan-id]",
		Short: "Fetch a premium plan by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			planID, err := parseUint64(args[0])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.PremiumPlan(cmd.Context(), &types.QueryPremiumPlanRequest{PlanId: planID})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func newQueryPlansCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plans",
		Short: "List premium plans with optional filters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			policyID, _ := cmd.Flags().GetUint64(flagPolicyID)
			payer, _ := cmd.Flags().GetString(flagPayer)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.PremiumPlans(cmd.Context(), &types.QueryPremiumPlansRequest{
				PolicyId:   policyID,
				Payer:      payer,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Uint64(flagPolicyID, 0, "Filter by policy ID")
	cmd.Flags().String(flagPayer, "", "Filter by payer address")
	flags.AddPaginationFlagsToCmd(cmd, "premium plans")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func newQueryPaymentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payments [plan-id]",
		Short: "List payments for a premium plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			planID, err := parseUint64(args[0])
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.PremiumPayments(cmd.Context(), &types.QueryPremiumPaymentsRequest{
				PlanId:     planID,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "premium payments")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func newQueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Show the current premium module parameters",
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

const (
	flagPolicyID = "policy-id"
	flagPayer    = "payer"
)
