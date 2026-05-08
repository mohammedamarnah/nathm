package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "1.0.0-dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version and exit",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), "nathm", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
