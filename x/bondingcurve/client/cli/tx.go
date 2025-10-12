package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/osmosis-labs/osmosis/v30/x/bondingcurve/types"
)

const (
	flagReason   = "reason"
	flagResumeIn = "resume-in"
	flagDuration = "duration"
	flagSigner   = "signer"
)

// GetTxCmd returns the root transaction command for the bondingcurve module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Bonding curve module transactions",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(
		newUpdateParamsCmd(),
		newSetEmergencyPauseCmd(),
		newSetTokenPauseCmd(),
		newForceLiquidationCmd(),
		newSetFreezeCmd(),
	)

	return cmd
}

func newUpdateParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-params [authority] [params-json]",
		Short: "Queue a parameter update via governance authority",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			filePath := args[1]

			contents, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read params file: %w", err)
			}

			var params types.Params
			if err := json.Unmarshal(contents, &params); err != nil {
				return fmt.Errorf("failed to unmarshal params: %w", err)
			}

			if err := params.Validate(); err != nil {
				return fmt.Errorf("invalid params: %w", err)
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgUpdateParams{Authority: authority, Params: params}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newSetEmergencyPauseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-emergency-pause [authority] [paused]",
		Short: "Toggle the global emergency pause",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			paused, err := strconv.ParseBool(args[1])
			if err != nil {
				return fmt.Errorf("invalid paused value: %w", err)
			}

			reason, _ := cmd.Flags().GetString(flagReason)
			resumeInStr, _ := cmd.Flags().GetString(flagResumeIn)
			signers, _ := cmd.Flags().GetStringSlice(flagSigner)
			if len(signers) == 0 {
				return fmt.Errorf("at least one signer must be provided via --%s", flagSigner)
			}

			var resumeDuration time.Duration
			if resumeInStr != "" {
				duration, err := time.ParseDuration(resumeInStr)
				if err != nil {
					return fmt.Errorf("invalid resume duration: %w", err)
				}
				if duration < 0 {
					return fmt.Errorf("resume duration cannot be negative")
				}
				resumeDuration = duration
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgSetEmergencyPause{
				Authority: authority,
				Paused:    paused,
				Reason:    reason,
				ResumeIn:  resumeDuration,
				Signers:   signers,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagReason, "", "Optional reason for the action")
	cmd.Flags().String(flagResumeIn, "", "Optional resume delay (e.g. 30m, 1h)")
	cmd.Flags().StringSlice(flagSigner, nil, "Addresses approving the emergency action")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newSetTokenPauseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-token-pause [authority] [denom] [paused]",
		Short: "Pause or resume a specific token",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			denom := args[1]
			paused, err := strconv.ParseBool(args[2])
			if err != nil {
				return fmt.Errorf("invalid paused value: %w", err)
			}

			reason, _ := cmd.Flags().GetString(flagReason)
			resumeInStr, _ := cmd.Flags().GetString(flagResumeIn)
			signers, _ := cmd.Flags().GetStringSlice(flagSigner)
			if len(signers) == 0 {
				return fmt.Errorf("at least one signer must be provided via --%s", flagSigner)
			}

			var resumeDuration time.Duration
			if resumeInStr != "" {
				duration, err := time.ParseDuration(resumeInStr)
				if err != nil {
					return fmt.Errorf("invalid resume duration: %w", err)
				}
				if duration < 0 {
					return fmt.Errorf("resume duration cannot be negative")
				}
				resumeDuration = duration
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgSetTokenPause{
				Authority: authority,
				Denom:     denom,
				Paused:    paused,
				Reason:    reason,
				ResumeIn:  resumeDuration,
				Signers:   signers,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagReason, "", "Optional reason for the action")
	cmd.Flags().String(flagResumeIn, "", "Optional resume delay (e.g. 30m, 1h)")
	cmd.Flags().StringSlice(flagSigner, nil, "Addresses approving the emergency action")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newForceLiquidationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "force-liquidation [authority] [position-id]...",
		Short: "Force liquidate specific margin positions",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			idArgs := args[1:]
			positionIDs := make([]uint64, 0, len(idArgs))
			for _, idStr := range idArgs {
				id, err := strconv.ParseUint(idStr, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid position id %q: %w", idStr, err)
				}
				positionIDs = append(positionIDs, id)
			}

			reason, _ := cmd.Flags().GetString(flagReason)
			signers, _ := cmd.Flags().GetStringSlice(flagSigner)
			if len(signers) == 0 {
				return fmt.Errorf("at least one signer must be provided via --%s", flagSigner)
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgForceLiquidation{
				Authority:   authority,
				PositionIds: positionIDs,
				Reason:      reason,
				Signers:     signers,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagReason, "", "Optional reason for the action")
	cmd.Flags().StringSlice(flagSigner, nil, "Addresses approving the emergency action")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newSetFreezeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-freeze [authority] [target-type] [target] [frozen]",
		Short: "Freeze or unfreeze a token or address",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			targetTypeStr := strings.ToLower(args[1])
			target := args[2]
			frozen, err := strconv.ParseBool(args[3])
			if err != nil {
				return fmt.Errorf("invalid frozen value: %w", err)
			}

			var targetType types.FreezeTargetType
			switch targetTypeStr {
			case "address", "addr":
				targetType = types.FreezeTargetType_FREEZE_TARGET_TYPE_ADDRESS
			case "token", "denom":
				targetType = types.FreezeTargetType_FREEZE_TARGET_TYPE_TOKEN
			default:
				return fmt.Errorf("unknown freeze target type %q", targetTypeStr)
			}

			reason, _ := cmd.Flags().GetString(flagReason)
			durationStr, _ := cmd.Flags().GetString(flagDuration)
			signers, _ := cmd.Flags().GetStringSlice(flagSigner)
			if len(signers) == 0 {
				return fmt.Errorf("at least one signer must be provided via --%s", flagSigner)
			}

			var freezeDuration time.Duration
			if durationStr != "" {
				duration, err := time.ParseDuration(durationStr)
				if err != nil {
					return fmt.Errorf("invalid duration: %w", err)
				}
				if duration < 0 {
					return fmt.Errorf("duration cannot be negative")
				}
				freezeDuration = duration
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgSetFreeze{
				Authority:  authority,
				TargetType: targetType,
				Target:     target,
				Frozen:     frozen,
				Reason:     reason,
				Duration:   freezeDuration,
				Signers:    signers,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagReason, "", "Optional reason for the action")
	cmd.Flags().String(flagDuration, "", "Optional freeze duration (e.g. 24h)")
	cmd.Flags().StringSlice(flagSigner, nil, "Addresses approving the emergency action")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
