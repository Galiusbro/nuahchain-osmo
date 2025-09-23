package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/premium/types"
)

// GetTxCmd returns the root transaction command for the premium module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "premium",
		Short:                      "Premium module transactions",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(
		newCreatePlanCmd(),
		newRecordPaymentCmd(),
		newMarkOverdueCmd(),
	)

	return cmd
}

func newCreatePlanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-plan [authority] [policy-id] [payer] [amount]",
		Short: "Create a premium plan for a policy",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			policyID, err := parseUint64(args[1])
			if err != nil {
				return err
			}
			payer := args[2]

			amount, err := sdk.ParseCoinNormalized(args[3])
			if err != nil {
				return fmt.Errorf("invalid amount: %w", err)
			}

			schedule, err := parseSchedule(cmd)
			if err != nil {
				return err
			}

			treasuryPoolID, err := cmd.Flags().GetString(flagTreasuryPool)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgCreatePremiumPlan{
				Authority:      authority,
				PolicyId:       policyID,
				Payer:          payer,
				Amount:         amount,
				Schedule:       schedule,
				TreasuryPoolId: treasuryPoolID,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Uint64(flagPeriod, 0, "Payment period in seconds")
	cmd.Flags().Uint64(flagMaxPayments, 0, "Maximum number of payments")
	cmd.Flags().String(flagScheduleType, "periodic", "Schedule type identifier")
	cmd.Flags().String(flagTreasuryPool, "", "Treasury pool ID")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func newRecordPaymentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record-payment [payer] [plan-id] [amount]",
		Short: "Record a premium payment",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			payer := args[0]
			planID, err := parseUint64(args[1])
			if err != nil {
				return err
			}
			amount, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return fmt.Errorf("invalid amount: %w", err)
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgRecordPremiumPayment{
				Payer:  payer,
				PlanId: planID,
				Amount: amount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newMarkOverdueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mark-overdue [authority] [plan-id]",
		Short: "Mark a premium plan as overdue",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			planID, err := parseUint64(args[1])
			if err != nil {
				return err
			}

			reason, err := cmd.Flags().GetString(flagReason)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgMarkPremiumOverdue{
				Authority: authority,
				PlanId:    planID,
				Reason:    reason,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagReason, "", "Overdue reason")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

const (
	flagPeriod       = "period"
	flagMaxPayments  = "max-payments"
	flagScheduleType = "schedule-type"
	flagTreasuryPool = "treasury-pool"
	flagReason       = "reason"
)

func parseSchedule(cmd *cobra.Command) (types.PremiumSchedule, error) {
	period, err := cmd.Flags().GetUint64(flagPeriod)
	if err != nil {
		return types.PremiumSchedule{}, err
	}
	maxPayments, err := cmd.Flags().GetUint64(flagMaxPayments)
	if err != nil {
		return types.PremiumSchedule{}, err
	}
	schedType, err := cmd.Flags().GetString(flagScheduleType)
	if err != nil {
		return types.PremiumSchedule{}, err
	}

	if period == 0 {
		return types.PremiumSchedule{}, fmt.Errorf("period must be > 0")
	}

	return types.PremiumSchedule{
		ScheduleType:  schedType,
		PeriodSeconds: period,
		MaxPayments:   maxPayments,
	}, nil
}

func parseUint64(input string) (uint64, error) {
	value, err := strconv.ParseUint(strings.TrimSpace(input), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid uint64: %s", input)
	}
	return value, nil
}
