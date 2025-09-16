package cli

import (
	"fmt"
	"strconv"
	"strings"

	"cosmossdk.io/math"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/osmosis-labs/osmosis/v30/x/exchange/types"
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

	cmd.AddCommand(CmdExchangeTokens())
	cmd.AddCommand(CmdUpdateParams())

	return cmd
}

// CmdExchangeTokens returns a CLI command handler for creating a MsgExchangeTokens transaction.
func CmdExchangeTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exchange-tokens [token-in] [min-nuah-out]",
		Short: "Exchange tokens for N$ (NUAH)",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Exchange supported tokens for N$ (NUAH) tokens.

Example:
$ %s tx %s exchange-tokens 1000000uosmo 950000000 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse token input
			tokenIn, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return fmt.Errorf("invalid token input: %w", err)
			}

			// Parse minimum NUAH output
			minNuahOut, ok := math.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid minimum NUAH output: %s", args[1])
			}

			msg := types.NewMsgExchangeTokens(
				clientCtx.GetFromAddress().String(),
				tokenIn,
				minNuahOut,
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

// CmdUpdateParams returns a CLI command handler for creating a MsgUpdateParams transaction.
func CmdUpdateParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-params [authority] [enabled] [admin] [min-exchange-amount-usd] [max-exchange-amount-usd] [daily-limit-usd] [exchange-fee] [treasury-addresses]",
		Short: "Update the exchange module parameters (governance only)",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Update the exchange module parameters. This command can only be executed via governance proposal.

Example:
$ %s tx %s update-params cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn true admin_address 10.0 100000.0 1000000.0 0.001 treasury1,treasury2 --from mykey
`,
				version.AppName, types.ModuleName,
			),
		),
		Args: cobra.ExactArgs(8),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			authority := args[0]

			// Parse enabled
			enabled, err := strconv.ParseBool(args[1])
			if err != nil {
				return fmt.Errorf("invalid enabled value: %w", err)
			}

			admin := args[2]

			// Parse min exchange amount USD
			minExchangeAmountUsd, err := math.LegacyNewDecFromStr(args[3])
			if err != nil {
				return fmt.Errorf("invalid min exchange amount USD: %w", err)
			}

			// Parse max exchange amount USD
			maxExchangeAmountUsd, err := math.LegacyNewDecFromStr(args[4])
			if err != nil {
				return fmt.Errorf("invalid max exchange amount USD: %w", err)
			}

			// Parse daily limit USD
			dailyLimitUsd, err := math.LegacyNewDecFromStr(args[5])
			if err != nil {
				return fmt.Errorf("invalid daily limit USD: %w", err)
			}

			// Parse exchange fee
			exchangeFee, err := math.LegacyNewDecFromStr(args[6])
			if err != nil {
				return fmt.Errorf("invalid exchange fee: %w", err)
			}

			// Parse treasury addresses
			var treasuryAddresses []string
			if args[7] != "" {
				treasuryAddresses = strings.Split(args[7], ",")
			}

			params := types.NewParams(
				enabled,
				admin,
				minExchangeAmountUsd,
				maxExchangeAmountUsd,
				dailyLimitUsd,
				math.LegacyNewDecWithPrec(2, 2), // priceDeviationThreshold - 2%
				exchangeFee,
				treasuryAddresses,
				[]string{}, // supportedTokens - empty by default
			)

			msg := types.NewMsgUpdateParams(authority, params)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}