package cli

import "github.com/spf13/cobra"

// GetTxCmd returns the root transaction command for the claims module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "claims",
		Short:                      "Claims module transactions",
		SilenceUsage:               true,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	return cmd
}
