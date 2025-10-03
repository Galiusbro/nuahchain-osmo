package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/leverage/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdOpenPosition())
	cmd.AddCommand(CmdClosePosition())
	cmd.AddCommand(CmdAddCollateral())
	cmd.AddCommand(CmdRemoveCollateral())
	cmd.AddCommand(CmdLiquidatePosition())
	cmd.AddCommand(CmdProvideLiquidity())

	return cmd
}

func CmdOpenPosition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-position [token-denom] [collateral-amount] [collateral-denom] [leverage] [side] [min-price] [max-price]",
		Short: "Open a new leverage position",
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			tokenDenom := args[0]

			collateralAmount, ok := math.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid collateral amount: %s", args[1])
			}

			collateralDenom := args[2]

			leverage, err := math.LegacyNewDecFromStr(args[3])
			if err != nil {
				return fmt.Errorf("invalid leverage: %w", err)
			}

			var side types.PositionSide
			switch args[4] {
			case "long", "LONG":
				side = types.PositionSideLong
			case "short", "SHORT":
				side = types.PositionSideShort
			default:
				return fmt.Errorf("invalid side: %s (must be 'long' or 'short')", args[4])
			}

			minPrice, err := math.LegacyNewDecFromStr(args[5])
			if err != nil {
				return fmt.Errorf("invalid min price: %w", err)
			}

			maxPrice, err := math.LegacyNewDecFromStr(args[6])
			if err != nil {
				return fmt.Errorf("invalid max price: %w", err)
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.MsgOpenPosition{
				Trader:     clientCtx.GetFromAddress().String(),
				TokenDenom: tokenDenom,
				Collateral: sdk.NewCoin(collateralDenom, collateralAmount),
				Leverage:   leverage,
				Side:       side,
				MinPrice:   minPrice,
				MaxPrice:   maxPrice,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdClosePosition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close-position [position-id] [min-price] [max-price]",
		Short: "Close an existing position",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			positionID := args[0]

			minPrice, err := math.LegacyNewDecFromStr(args[1])
			if err != nil {
				return fmt.Errorf("invalid min price: %w", err)
			}

			maxPrice, err := math.LegacyNewDecFromStr(args[2])
			if err != nil {
				return fmt.Errorf("invalid max price: %w", err)
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.MsgClosePosition{
				Trader:     clientCtx.GetFromAddress().String(),
				PositionId: positionID,
				MinPrice:   minPrice,
				MaxPrice:   maxPrice,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdAddCollateral() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-collateral [position-id] [amount] [denom]",
		Short: "Add collateral to an existing position",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			positionID := args[0]

			amount, ok := math.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid amount: %s", args[1])
			}

			denom := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.MsgAddCollateral{
				Trader:     clientCtx.GetFromAddress().String(),
				PositionId: positionID,
				Amount:     sdk.NewCoin(denom, amount),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdRemoveCollateral() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-collateral [position-id] [amount] [denom]",
		Short: "Remove collateral from an existing position",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			positionID := args[0]

			amount, ok := math.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid amount: %s", args[1])
			}

			denom := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.MsgRemoveCollateral{
				Trader:     clientCtx.GetFromAddress().String(),
				PositionId: positionID,
				Amount:     sdk.NewCoin(denom, amount),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdLiquidatePosition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquidate-position [position-id]",
		Short: "Liquidate a position that has reached liquidation threshold",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			positionID := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.MsgLiquidatePosition{
				Liquidator: clientCtx.GetFromAddress().String(),
				PositionId: positionID,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdProvideLiquidity() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provide-liquidity [amount] [denom]",
		Short: "Provide liquidity to a lending pool",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			amount, ok := math.NewIntFromString(args[0])
			if !ok {
				return fmt.Errorf("invalid amount: %s", args[0])
			}

			denom := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.MsgProvideLiquidity{
				Provider: clientCtx.GetFromAddress().String(),
				Amount:   sdk.NewCoin(denom, amount),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
