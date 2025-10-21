package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/osmosis-labs/osmosis/v30/x/oracle/types"
)

// GetTxCmd returns the root tx command for the oracle module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Oracle transactions subcommands",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(newSetPriceCmd())

	return cmd
}

func newSetPriceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-price [symbol] [value]",
		Short: "Set the price for a symbol",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			symbol := strings.TrimSpace(args[0])
			value := strings.TrimSpace(args[1])

			msg := types.NewMsgSetPrice(clientCtx.GetFromAddress().String(), symbol, value)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
