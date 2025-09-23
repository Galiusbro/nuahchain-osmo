package cli

import "github.com/spf13/cobra"

// GetQueryCmd returns the root query command for the premium module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "premium",
		Short:                      "Query commands for the premium module",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	return cmd
}
