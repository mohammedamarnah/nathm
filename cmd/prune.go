package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/USER/nathm/internal/branch"
	"github.com/USER/nathm/internal/config"
	"github.com/USER/nathm/internal/git"
)

var pruneYes bool

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Delete branches whose upstream is gone or which are merged into base",
	RunE:  runPrune,
}

func runPrune(cmd *cobra.Command, args []string) error {
	g := git.NewExec("")
	if !g.IsRepo() {
		return fmt.Errorf("not a git repository")
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := g.FetchPrune(); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", err)
	}
	bs, err := branch.Load(g, branch.LoadConfig{
		BaseBranches:      cfg.BaseBranches,
		ProtectedPatterns: cfg.ProtectedPatterns,
	})
	if err != nil {
		return err
	}
	var candidates []branch.Branch
	for _, b := range bs {
		if b.IsStale() && !b.Protected {
			candidates = append(candidates, b)
		}
	}
	if len(candidates) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No stale branches to prune.")
		return nil
	}
	printCandidates(cmd.OutOrStdout(), candidates)

	if !pruneYes {
		ok, err := confirm(cmd.InOrStdin(), cmd.OutOrStdout())
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
			os.Exit(2)
		}
	}

	var deleted, failed int
	for _, b := range candidates {
		if err := g.DeleteBranch(b.Name, false); err != nil {
			if b.UpstreamGone {
				if err2 := g.DeleteBranch(b.Name, true); err2 == nil {
					deleted++
					fmt.Fprintf(cmd.OutOrStdout(), "deleted %s (force)\n", b.Name)
					continue
				}
			}
			failed++
			fmt.Fprintf(cmd.ErrOrStderr(), "failed: %s: %v\n", b.Name, err)
			continue
		}
		deleted++
		fmt.Fprintf(cmd.OutOrStdout(), "deleted %s\n", b.Name)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "\nDone: %d deleted, %d failed.\n", deleted, failed)
	if deleted == 0 && failed > 0 {
		return fmt.Errorf("all deletions failed")
	}
	return nil
}

func printCandidates(w io.Writer, candidates []branch.Branch) {
	fmt.Fprintf(w, "The following %d branch(es) will be deleted:\n", len(candidates))
	now := time.Now()
	for _, b := range candidates {
		age := humanize.RelTime(b.LastCommitTime, now, "ago", "from now")
		fmt.Fprintf(w, "  %-30s %-12s last commit %s\n", b.Name, b.Status(), age)
	}
}

func confirm(in io.Reader, out io.Writer) (bool, error) {
	fmt.Fprint(out, "Continue? [y/N] ")
	r := bufio.NewReader(in)
	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes", nil
}

func init() {
	pruneCmd.Flags().BoolVar(&pruneYes, "yes", false, "Skip confirmation prompt")
	rootCmd.AddCommand(pruneCmd)
}
