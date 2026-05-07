package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/USER/nathm/internal/branch"
	"github.com/USER/nathm/internal/config"
	"github.com/USER/nathm/internal/git"
)

var listStaleOnly bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Print local branches as TSV (name, status, age_seconds, ahead, behind)",
	RunE: func(cmd *cobra.Command, args []string) error {
		g := git.NewExec("")
		if !g.IsRepo() {
			return fmt.Errorf("not a git repository")
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		bs, err := branch.Load(g, branch.LoadConfig{
			BaseBranches:      cfg.BaseBranches,
			ProtectedPatterns: cfg.ProtectedPatterns,
		})
		if err != nil {
			return err
		}
		now := time.Now()
		w := cmd.OutOrStdout()
		for _, b := range bs {
			if listStaleOnly && !b.IsStale() {
				continue
			}
			age := int64(now.Sub(b.LastCommitTime).Seconds())
			fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\n", b.Name, b.Status(), age, b.Ahead, b.Behind)
		}
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVar(&listStaleOnly, "stale", false, "Only print stale (gone or merged) branches")
	rootCmd.AddCommand(listCmd)
}
