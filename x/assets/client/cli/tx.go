package cli

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

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
