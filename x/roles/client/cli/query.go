package cli

import "github.com/spf13/cobra"

// GetQueryCmd returns the root query command for the roles module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "roles",
		Short:                      "Query commands for the roles module",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	return cmd
}
