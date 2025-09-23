package cli

import "github.com/spf13/cobra"

// GetQueryCmd returns the root query command for the policy module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "policy",
		Short:                      "Query commands for the policy module",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	return cmd
}
