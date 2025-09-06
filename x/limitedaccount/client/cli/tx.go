package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	limitedaccounttypes "github.com/osmosis-labs/osmosis/v30/x/limitedaccount/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        limitedaccounttypes.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", limitedaccounttypes.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdCreateLimitedAccount())

	return cmd
}

// CmdCreateLimitedAccount returns a CLI command handler for creating a limited account
func CmdCreateLimitedAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-limited-account [address]",
		Short: "Create a limited account with 3 free transactions per day",
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

			msg := &limitedaccounttypes.MsgCreateLimitedAccount{
				Authority: authority,
				Address:   args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
