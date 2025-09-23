package cli

import "github.com/spf13/cobra"

// GetQueryCmd returns the root query command for the treasury module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "treasury",
		Short:                      "Query commands for the treasury module",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	return cmd
}
