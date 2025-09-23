package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v30/x/claims/types"
)

// GetTxCmd returns the root transaction command for the claims module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "claims",
		Short:                      "Claims module transactions",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(
		newSubmitClaimCmd(),
		newReviewClaimCmd(),
		newAddEvidenceCmd(),
		newExecutePayoutCmd(),
	)

	return cmd
}

func newSubmitClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit [claimant] [policy-id] [amount]",
		Short: "Submit a new claim",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			claimant := args[0]
			policyID, err := parseUint64(args[1])
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return fmt.Errorf("invalid amount: %w", err)
			}

			description, _ := cmd.Flags().GetString(flagDescription)
			evidenceURIs, _ := cmd.Flags().GetStringSlice(flagEvidence)

			evidences := make([]types.ClaimEvidence, 0, len(evidenceURIs))
			for _, uri := range evidenceURIs {
				evidences = append(evidences, types.ClaimEvidence{Uri: uri})
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgSubmitClaim{
				Claimant:    claimant,
				PolicyId:    policyID,
				Amount:      amount,
				Description: description,
				Evidence:    evidences,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagDescription, "", "Claim description")
	cmd.Flags().StringSlice(flagEvidence, nil, "Evidence URIs (repeatable)")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newReviewClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review [authority] [claim-id] [decision]",
		Short: "Review a claim decision",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			claimID, err := parseUint64(args[1])
			if err != nil {
				return err
			}

			decision, err := parseClaimStatus(args[2])
			if err != nil {
				return err
			}

			reason, _ := cmd.Flags().GetString(flagReason)

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgReviewClaim{
				Authority: authority,
				ClaimId:   claimID,
				Decision:  decision,
				Reason:    reason,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagReason, "", "Decision rationale")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newAddEvidenceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-evidence [authority] [claim-id] [uri]",
		Short: "Attach evidence to a claim",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			claimID, err := parseUint64(args[1])
			if err != nil {
				return err
			}

			uri := args[2]
			notes, _ := cmd.Flags().GetString(flagNotes)

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgAddClaimEvidence{
				Authority: authority,
				ClaimId:   claimID,
				Evidence: types.ClaimEvidence{
					Uri:   uri,
					Notes: notes,
				},
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagNotes, "", "Evidence notes")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newExecutePayoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute-payout [authority] [claim-id] [recipient]",
		Short: "Execute payout for an approved claim",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			claimID, err := parseUint64(args[1])
			if err != nil {
				return err
			}
			recipient := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgExecuteClaimPayout{
				Authority:     authority,
				ClaimId:       claimID,
				PayoutAddress: recipient,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

const (
	flagDescription = "description"
	flagEvidence    = "evidence"
	flagReason      = "reason"
	flagNotes       = "notes"
	flagPolicyID    = "policy-id"
	flagClaimant    = "claimant"
	flagStatus      = "status"
)

func parseClaimStatus(input string) (types.ClaimStatus, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "pending":
		return types.ClaimStatus_CLAIM_STATUS_PENDING, nil
	case "approved":
		return types.ClaimStatus_CLAIM_STATUS_APPROVED, nil
	case "rejected":
		return types.ClaimStatus_CLAIM_STATUS_REJECTED, nil
	default:
		return types.ClaimStatus_CLAIM_STATUS_UNSPECIFIED, fmt.Errorf("unknown decision: %s", input)
	}
}

func parseUint64(input string) (uint64, error) {
	value, err := strconv.ParseUint(strings.TrimSpace(input), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid uint64: %s", input)
	}
	return value, nil
}
