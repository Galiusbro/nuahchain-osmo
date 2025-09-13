package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/osmosis-labs/osmosis/v30/x/usdoracle/types"
)

var DefaultRelativePacketTimeoutTimestamp = uint64((20 * 60 * 1000000000)) // 20 minutes

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
	cmd.AddCommand(CmdUpdateUSDPrice())
	cmd.AddCommand(CmdSetPriceSources())

	return cmd
}

func CmdUpdateParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-params [enabled] [admin] [update-interval] [price-deviation-threshold]",
		Short: "Update USD oracle module parameters",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			enabled, err := strconv.ParseBool(args[0])
			if err != nil {
				return err
			}

			admin := args[1]

			updateInterval, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			priceDeviationThreshold, err := math.LegacyNewDecFromStr(args[3])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			params := types.NewParams(enabled, admin, updateInterval, priceDeviationThreshold)

			msg := &types.MsgUpdateParams{
				Authority: clientCtx.GetFromAddress().String(),
				Params:    params,
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

func CmdUpdateUSDPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-usd-price [price] [source]",
		Short: "Manually update USD price",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			price, err := math.LegacyNewDecFromStr(args[0])
			if err != nil {
				return err
			}

			source := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgUpdateUSDPrice{
				Authority: clientCtx.GetFromAddress().String(),
				Price:     price,
				Source:    source,
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

func CmdSetPriceSources() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-price-sources [sources]",
		Short: "Set price sources configuration",
		Long: `Set price sources configuration. Sources should be provided as comma-separated list in format:
name1:weight1:enabled1:url1,name2:weight2:enabled2:url2

Example:
coingecko:0.6:true:https://api.coingecko.com/api/v3/simple/price,binance:0.4:true:https://api.binance.com/api/v3/ticker/price`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			sourcesStr := args[0]
			sourcesList := strings.Split(sourcesStr, ",")

			var sources []types.PriceSource
			for _, sourceStr := range sourcesList {
				parts := strings.Split(sourceStr, ":")
				if len(parts) != 4 {
					return fmt.Errorf("invalid source format: %s. Expected name:weight:enabled:url", sourceStr)
				}

				name := parts[0]
				weight, err := math.LegacyNewDecFromStr(parts[1])
				if err != nil {
					return fmt.Errorf("invalid weight for source %s: %w", name, err)
				}

				enabled, err := strconv.ParseBool(parts[2])
				if err != nil {
					return fmt.Errorf("invalid enabled flag for source %s: %w", name, err)
				}

				url := parts[3]

				sources = append(sources, types.PriceSource{
					Name:    name,
					Weight:  weight,
					Enabled: enabled,
					Url:     url,
				})
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgSetPriceSources{
				Authority: clientCtx.GetFromAddress().String(),
				Sources:   sources,
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
