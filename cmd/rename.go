package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mohammedamarnah/nathm/internal/git"
)

var renameCmd = &cobra.Command{
	Use:   "rename <old> <new>",
	Short: "Rename a local branch",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		g := git.NewExec("")
		if !g.IsRepo() {
			return fmt.Errorf("not a git repository")
		}
		return g.RenameBranch(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)
}
