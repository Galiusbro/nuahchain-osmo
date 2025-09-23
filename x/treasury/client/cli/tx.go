package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/treasury/types"
)

// GetTxCmd returns the root transaction command for the treasury module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "treasury",
		Short:                      "Treasury module transactions",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(
		newCreatePoolCmd(),
		newUpdatePoolCmd(),
		newDepositCmd(),
		newWithdrawCmd(),
		newSetReservesCmd(),
	)

	return cmd
}

func newCreatePoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool [authority] [pool-id]",
		Short: "Create a treasury pool",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			poolID := args[1]

			description, _ := cmd.Flags().GetString(flagDescription)
			manager, _ := cmd.Flags().GetString(flagManager)
			policyTypes, _ := cmd.Flags().GetStringSlice(flagPolicyTypes)

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgCreateTreasuryPool{
				Authority:   authority,
				PoolId:      poolID,
				Description: description,
				Manager:     manager,
				PolicyTypes: policyTypes,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagDescription, "", "Pool description")
	cmd.Flags().String(flagManager, "", "Pool manager address")
	cmd.Flags().StringSlice(flagPolicyTypes, nil, "Allowed policy types")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newUpdatePoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-pool [authority] [pool-id]",
		Short: "Update treasury pool metadata",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			poolID := args[1]

			description, _ := cmd.Flags().GetString(flagDescription)
			manager, _ := cmd.Flags().GetString(flagManager)
			policyTypes, _ := cmd.Flags().GetStringSlice(flagPolicyTypes)

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgUpdateTreasuryPool{
				Authority:   authority,
				PoolId:      poolID,
				Description: description,
				Manager:     manager,
				PolicyTypes: policyTypes,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagDescription, "", "Pool description")
	cmd.Flags().String(flagManager, "", "Pool manager address")
	cmd.Flags().StringSlice(flagPolicyTypes, nil, "Allowed policy types")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newDepositCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [depositor] [pool-id] [amount]",
		Short: "Deposit into treasury pool",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			depositor := args[0]
			poolID := args[1]
			amount, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return fmt.Errorf("invalid amount: %w", err)
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgDepositToTreasury{
				Depositor: depositor,
				PoolId:    poolID,
				Amount:    amount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newWithdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw [authority] [pool-id] [recipient] [amount]",
		Short: "Withdraw funds from treasury pool",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			poolID := args[1]
			recipient := args[2]
			amount, err := sdk.ParseCoinNormalized(args[3])
			if err != nil {
				return fmt.Errorf("invalid amount: %w", err)
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgWithdrawFromTreasury{
				Authority: authority,
				PoolId:    poolID,
				Recipient: recipient,
				Amount:    amount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newSetReservesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-reserves [authority] [pool-id] [denom=ratio ...]",
		Short: "Set reserve targets for a pool",
		Args:  cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			poolID := args[1]
			pairs := args[2:]

			reserves := make([]types.PoolReserves, 0, len(pairs))
			for _, pair := range pairs {
				denom, ratio, found := strings.Cut(pair, "=")
				if !found || denom == "" || ratio == "" {
					return fmt.Errorf("invalid reserve format: %s", pair)
				}
				reserves = append(reserves, types.PoolReserves{PoolId: poolID, Denom: denom, MinReserveRatio: ratio})
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgSetPoolReserves{
				Authority: authority,
				PoolId:    poolID,
				Reserves:  reserves,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

const (
	flagDescription = "description"
	flagManager     = "manager"
	flagPolicyTypes = "policy-type"
)
