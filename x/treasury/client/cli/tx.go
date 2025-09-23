package cli

import "github.com/spf13/cobra"

// GetTxCmd returns the root transaction command for the treasury module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "treasury",
		Short:                      "Treasury module transactions",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	return cmd
}
