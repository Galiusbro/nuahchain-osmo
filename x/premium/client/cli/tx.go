package cli

import "github.com/spf13/cobra"

// GetTxCmd returns the root transaction command for the premium module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "premium",
		Short:                      "Premium module transactions",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	return cmd
}
