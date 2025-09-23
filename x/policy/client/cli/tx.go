package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/osmosis-labs/osmosis/v30/x/policy/types"
)

// GetTxCmd returns the root transaction command for the policy module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "policy",
		Short:                      "Policy module transactions",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(
		newCreatePolicyCmd(),
		newUpdateAttributesCmd(),
		newCancelPolicyCmd(),
		newUpdateStatusCmd(),
	)

	return cmd
}

func newCreatePolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [owner] [policy-type]",
		Short: "Create a new policy",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			owner := args[0]
			policyType := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			attrPairs, err := cmd.Flags().GetStringArray(flagAttribute)
			if err != nil {
				return err
			}

			attributes, err := parseAttributes(attrPairs)
			if err != nil {
				return err
			}

			startUnix, err := cmd.Flags().GetInt64(flagStart)
			if err != nil {
				return err
			}

			endUnix, err := cmd.Flags().GetInt64(flagEnd)
			if err != nil {
				return err
			}

			treasuryPoolID, err := cmd.Flags().GetString(flagTreasuryPool)
			if err != nil {
				return err
			}

			tags, err := cmd.Flags().GetStringSlice(flagTags)
			if err != nil {
				return err
			}

			msg := &types.MsgCreatePolicy{
				Owner:          owner,
				PolicyType:     policyType,
				Attributes:     attributes,
				TreasuryPoolId: treasuryPoolID,
				Tags:           tags,
			}

			if startUnix > 0 {
				start := time.Unix(startUnix, 0).UTC()
				msg.StartTime = &start
			}
			if endUnix > 0 {
				end := time.Unix(endUnix, 0).UTC()
				msg.EndTime = &end
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().StringArray(flagAttribute, nil, "Key/Value attributes (key=value)")
	cmd.Flags().Int64(flagStart, 0, "Start time (unix seconds)")
	cmd.Flags().Int64(flagEnd, 0, "End time (unix seconds)")
	cmd.Flags().String(flagTreasuryPool, "", "Treasury pool ID")
	cmd.Flags().StringSlice(flagTags, nil, "Comma-separated policy tags")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func newUpdateAttributesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-attributes [authority] [policy-id]",
		Short: "Update policy attributes",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			policyID, err := parseUint64(args[1])
			if err != nil {
				return err
			}

			attrPairs, err := cmd.Flags().GetStringArray(flagAttribute)
			if err != nil {
				return err
			}

			attributes, err := parseAttributes(attrPairs)
			if err != nil {
				return err
			}

			replace, err := cmd.Flags().GetBool(flagReplace)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgUpdatePolicyAttributes{
				Authority:  authority,
				PolicyId:   policyID,
				Attributes: attributes,
				Replace:    replace,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().StringArray(flagAttribute, nil, "Key/Value attributes (key=value)")
	cmd.Flags().Bool(flagReplace, false, "Replace existing attributes instead of merging")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newCancelPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel [authority] [policy-id]",
		Short: "Cancel an existing policy",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			policyID, err := parseUint64(args[1])
			if err != nil {
				return err
			}

			reason, err := cmd.Flags().GetString(flagReason)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgCancelPolicy{
				Authority: authority,
				PolicyId:  policyID,
				Reason:    reason,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagReason, "", "Cancellation reason")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newUpdateStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-status [authority] [policy-id] [status]",
		Short: "Update policy status",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			policyID, err := parseUint64(args[1])
			if err != nil {
				return err
			}

			status, err := parseStatus(args[2])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgUpdatePolicyStatus{
				Authority: authority,
				PolicyId:  policyID,
				Status:    status,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

const (
	flagAttribute    = "attribute"
	flagStart        = "start"
	flagEnd          = "end"
	flagTreasuryPool = "treasury-pool"
	flagTags         = "tags"
	flagReplace      = "replace"
	flagReason       = "reason"
	flagOwner        = "owner"
	flagPolicyType   = "policy-type"
	flagStatus       = "status"
)

func parseUint64(input string) (uint64, error) {
	value, err := strconv.ParseUint(strings.TrimSpace(input), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid uint64: %s", input)
	}
	return value, nil
}

func parseAttributes(pairs []string) ([]types.PolicyAttribute, error) {
	if len(pairs) == 0 {
		return nil, nil
	}

	attrs := make([]types.PolicyAttribute, 0, len(pairs))
	for _, pair := range pairs {
		key, value, found := strings.Cut(pair, "=")
		if !found || key == "" {
			return nil, fmt.Errorf("invalid attribute format: %s", pair)
		}
		attrs = append(attrs, types.PolicyAttribute{Key: key, Value: value})
	}

	return attrs, nil
}

func parseStatus(input string) (types.PolicyStatus, error) {
	normalized := strings.ToLower(strings.TrimSpace(input))
	switch normalized {
	case "active":
		return types.PolicyStatus_POLICY_STATUS_ACTIVE, nil
	case "expired":
		return types.PolicyStatus_POLICY_STATUS_EXPIRED, nil
	case "claimed":
		return types.PolicyStatus_POLICY_STATUS_CLAIMED, nil
	case "cancelled", "canceled":
		return types.PolicyStatus_POLICY_STATUS_CANCELLED, nil
	default:
		return types.PolicyStatus_POLICY_STATUS_UNSPECIFIED, fmt.Errorf("unknown status: %s", input)
	}
}
