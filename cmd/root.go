package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nathm",
	Short: "Organize and clean up local git branches",
	Long:  "nathm (نظم) is an interactive TUI and CLI for organizing local git branches.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TUI wired up in Task 23. For now, print a placeholder.
		fmt.Fprintln(cmd.OutOrStdout(), "TUI not yet wired — try `nathm version` or `nathm --help`.")
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
