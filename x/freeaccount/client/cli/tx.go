package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	freeaccounttypes "github.com/osmosis-labs/osmosis/v30/x/freeaccount/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        freeaccounttypes.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", freeaccounttypes.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdCreateFreeAccount())

	return cmd
}

// CmdCreateFreeAccount returns a CLI command handler for creating a free account
func CmdCreateFreeAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-free-account [address]",
		Short: "Create a free account that doesn't pay transaction fees",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Validate address
			_, err = sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("invalid address: %w", err)
			}

			// Get authority (should be gov module account)
			authority := clientCtx.GetFromAddress().String()

			msg := freeaccounttypes.NewMsgCreateFreeAccount(authority, args[0])
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
