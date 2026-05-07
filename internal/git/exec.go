package git

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/mohammedamarnah/nathm/internal/branch"
)

// Exec is the real git CLI wrapper.
type Exec struct {
	dir string // working directory; "" means the process cwd
}

// NewExec returns an Exec wrapper rooted at dir. Pass "" to use the current
// working directory.
func NewExec(dir string) *Exec {
	return &Exec{dir: dir}
}

func (e *Exec) IsRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = e.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return string(out) == "true\n"
}

const forEachRefFormat = "%(HEAD)%00%(refname:short)%00%(upstream:short)%00%(upstream:track)%00%(committerdate:unix)%00%(objectname)%00%(contents:subject)"

func (e *Exec) ListBranches() ([]branch.Branch, error) {
	cmd := exec.Command("git", "for-each-ref", "--format="+forEachRefFormat, "refs/heads")
	cmd.Dir = e.dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git for-each-ref: %w", err)
	}
	return branch.ParseForEachRef(out)
}

func (e *Exec) AheadBehind(br, base string) (int, int, error) {
	cmd := exec.Command("git", "rev-list", "--left-right", "--count", base+"..."+br)
	cmd.Dir = e.dir
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("git rev-list: %w", err)
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected rev-list output: %q", out)
	}
	behind, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse behind: %w", err)
	}
	ahead, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse ahead: %w", err)
	}
	return ahead, behind, nil
}

func (e *Exec) MergedInto(br, base string) (bool, error) {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", br, base)
	cmd.Dir = e.dir
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		return false, nil // documented "not an ancestor" exit code
	}
	return false, fmt.Errorf("git merge-base: %w", err)
}

func (e *Exec) DeleteBranch(name string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	cmd := exec.Command("git", "branch", flag, name)
	cmd.Dir = e.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git branch %s %s: %w: %s", flag, name, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (e *Exec) RenameBranch(oldName, newName string) error {
	cmd := exec.Command("git", "branch", "-m", oldName, newName)
	cmd.Dir = e.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git branch -m: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (e *Exec) Checkout(name string) error {
	cmd := exec.Command("git", "checkout", "-q", name)
	cmd.Dir = e.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (e *Exec) FetchPrune() error {
	cmd := exec.Command("git", "fetch", "--all", "--prune", "-q")
	cmd.Dir = e.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch --prune: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Compile-time check that *Exec satisfies the Git interface.
var _ Git = (*Exec)(nil)
