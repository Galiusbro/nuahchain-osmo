package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/osmosis-labs/osmosis/v30/x/roles/types"
)

// GetTxCmd returns the root transaction command for the roles module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "roles",
		Short:                      "Roles module transactions",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	cmd.AddCommand(
		newAssignRolesCmd(),
		newRevokeRolesCmd(),
		newUpdateAuthorityCmd(),
	)

	return cmd
}

func newAssignRolesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assign [authority] [address] [role ...]",
		Short: "Assign one or more roles to an address",
		Args:  cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			address := args[1]
			roleStrs := args[2:]

			roles, err := parseRoleStrings(roleStrs)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgAssignRoles{
				Authority: authority,
				Address:   address,
				Roles:     roles,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newRevokeRolesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke [authority] [address] [role ...]",
		Short: "Revoke one or more roles from an address",
		Args:  cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			address := args[1]
			roleStrs := args[2:]

			roles, err := parseRoleStrings(roleStrs)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgRevokeRoles{
				Authority: authority,
				Address:   address,
				Roles:     roles,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newUpdateAuthorityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-authority [authority] [new-authority]",
		Short: "Transfer module authority to a new address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			authority := args[0]
			newAuthority := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgUpdateAuthority{
				Authority:    authority,
				NewAuthority: newAuthority,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func parseRoleStrings(inputs []string) ([]types.Role, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("at least one role must be provided")
	}

	roles := make([]types.Role, 0, len(inputs))
	for _, input := range inputs {
		role, err := roleFromString(input)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func roleFromString(value string) (types.Role, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "insurer", "role_insurer":
		return types.Role_ROLE_INSURER, nil
	case "policy_holder", "policy-holder", "holder", "role_policy_holder":
		return types.Role_ROLE_POLICY_HOLDER, nil
	case "claims_reviewer", "claims-reviewer", "reviewer", "role_claims_reviewer":
		return types.Role_ROLE_CLAIMS_REVIEWER, nil
	case "oracle", "role_oracle":
		return types.Role_ROLE_ORACLE, nil
	case "treasury_manager", "treasury-manager", "treasury", "role_treasury_manager":
		return types.Role_ROLE_TREASURY_MANAGER, nil
	default:
		return types.Role_ROLE_UNSPECIFIED, fmt.Errorf("unknown role: %s", value)
	}
}
