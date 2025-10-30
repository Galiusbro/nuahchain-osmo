package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	osmomath "github.com/osmosis-labs/osmosis/osmomath"

	"github.com/osmosis-labs/osmosis/v30/x/assets/types"
)

// GetTxCmd returns the root tx command for the assets module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Assets transactions subcommands",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(newEnsureAssetCmd())
	cmd.AddCommand(newBuyAssetCmd())
	cmd.AddCommand(newSellAssetCmd())

	return cmd
}

func newEnsureAssetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ensure [symbol]",
		Short: "Ensure that an asset exists, creating it if necessary",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			symbol := strings.TrimSpace(args[0])
			msg := types.NewMsgEnsureAsset(clientCtx.GetFromAddress().String(), symbol)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newBuyAssetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buy [symbol] [coin]",
		Short: "Buy an asset using NDOLLAR or unuah (e.g., 1000NDOLLAR or 1000unuah)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			symbol := strings.TrimSpace(args[0])
			coin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid coin format: %w (expected format: amount+denom, e.g., 1000NDOLLAR or 1000unuah)", err)
			}

			// Support both NDOLLAR and unuah (including factory/*/ndollar format)
			isNDollar := strings.EqualFold(coin.Denom, types.NDollarDenom) || (strings.HasPrefix(coin.Denom, "factory/") && strings.HasSuffix(coin.Denom, "/ndollar"))
			if !isNDollar && !strings.EqualFold(coin.Denom, "unuah") {
				return fmt.Errorf("denom must be %s (or factory/*/ndollar) or unuah, got %s", types.NDollarDenom, coin.Denom)
			}

			// Create message with new format (denom + amount)
			msg := &types.MsgBuyAsset{
				Buyer:  clientCtx.GetFromAddress().String(),
				Symbol: symbol,
				Denom:  coin.Denom,
				Amount: coin.Amount.String(),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newSellAssetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sell [symbol] [base_amount]",
		Short: "Sell an asset back for NDOLLAR",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			symbol := strings.TrimSpace(args[0])
			baseAmount := strings.TrimSpace(args[1])
			if _, err := osmomath.NewDecFromStr(baseAmount); err != nil {
				return fmt.Errorf("invalid base amount: %w", err)
			}

			msg := types.NewMsgSellAsset(clientCtx.GetFromAddress().String(), symbol, baseAmount)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
