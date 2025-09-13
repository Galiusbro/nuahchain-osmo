package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/osmosis-labs/osmosis/v30/x/pegkeeper/types"
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

	cmd.AddCommand(CmdUpdateParams())

	return cmd
}

// CmdUpdateParams returns a CLI command handler for updating pegkeeper parameters
func CmdUpdateParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-params [target-denom] [reference-denom] [max-deviation-threshold] [adjustment-factor] [min-adjustment-interval] [max-supply-change] [oracle-module] [enabled] [target-price]",
		Short: "Update pegkeeper parameters",
		Args:  cobra.ExactArgs(9),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			minAdjustmentInterval, err := strconv.ParseInt(args[4], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid min-adjustment-interval: %w", err)
			}

			enabled, err := strconv.ParseBool(args[7])
			if err != nil {
				return fmt.Errorf("invalid enabled flag: %w", err)
			}

			params := types.Params{
				MaxDeviationThreshold:        args[2],
				AdjustmentFactor:             args[3],
				MinAdjustmentInterval:        minAdjustmentInterval,
				MaxSupplyChangePerAdjustment: args[5],
				OracleModule:                 args[6],
				Enabled:                      enabled,
				TargetDenom:                  args[0],
				ReferenceDenom:               args[1],
				TargetPrice:                  args[8],
			}

			msg := types.NewMsgUpdateParams(clientCtx.GetFromAddress().String(), params)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
