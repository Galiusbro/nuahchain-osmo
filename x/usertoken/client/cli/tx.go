package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/usertoken/types"
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

	cmd.AddCommand(CmdCreateUserToken())
	cmd.AddCommand(CmdBuyTokens())
	cmd.AddCommand(CmdSellTokens())
	cmd.AddCommand(CmdClaimFounderTokens())
	cmd.AddCommand(CmdStartLBP())
	cmd.AddCommand(CmdCreateVestingAccount())

	return cmd
}

func CmdCreateUserToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-user-token [subdenom] [name] [symbol] [decimals]",
		Short: "Create a new user token",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			decimals, err := strconv.ParseUint(args[3], 10, 32)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateUserToken(
				clientCtx.GetFromAddress().String(),
				args[0],          // subdenom
				args[1],          // name
				args[2],          // symbol
				uint32(decimals), // decimals
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdCreateVestingAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-vesting-account [to-address] [amount] [end-time] [delayed]",
		Short: "Create a vesting account for token distribution",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse amount
			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			// Parse end time (Unix timestamp)
			endTime, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid end time: %w", err)
			}

			// Parse delayed flag
			delayed, err := strconv.ParseBool(args[3])
			if err != nil {
				return fmt.Errorf("invalid delayed flag: %w", err)
			}

			msg := types.NewMsgCreateVestingAccount(
				clientCtx.GetFromAddress().String(),
				args[0], // to_address
				amount,  // amount
				endTime, // end_time
				delayed, // delayed
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdBuyTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buy-tokens [denom] [amount] [min-tokens]",
		Short: "Buy tokens via bonding curve",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgBuyTokens(
				clientCtx.GetFromAddress().String(),
				args[0], // denom
				amount,  // amount
				args[2], // min_tokens
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdSellTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sell-tokens [denom] [amount] [min-price]",
		Short: "Sell tokens via bonding curve",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgSellTokens(
				clientCtx.GetFromAddress().String(),
				args[0], // denom
				amount,  // amount
				args[2], // min_price
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdClaimFounderTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-founder-tokens [denom] [amount]",
		Short: "Claim founder tokens at fixed price",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgClaimFounderTokens(
				clientCtx.GetFromAddress().String(),
				args[0], // denom
				args[1], // amount
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdStartLBP() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start-lbp [denom]",
		Short: "Start Liquidity Bootstrapping Pool for token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgStartLBP(
				clientCtx.GetFromAddress().String(),
				args[0], // denom
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
