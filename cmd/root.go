package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/USER/nathm/internal/branch"
	"github.com/USER/nathm/internal/config"
	"github.com/USER/nathm/internal/git"
	"github.com/USER/nathm/internal/tui"
)

var rootCmd = &cobra.Command{
	Use:   "nathm",
	Short: "Organize and clean up local git branches",
	Long:  "nathm (نظم) is an interactive TUI and CLI for organizing local git branches.",
	RunE:  runRoot,
}

func runRoot(cmd *cobra.Command, args []string) error {
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
	m := tui.NewModel(bs, g)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
