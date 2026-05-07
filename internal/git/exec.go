package git

import (
	"fmt"
	"os/exec"

	"github.com/USER/nathm/internal/branch"
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

// Compile-time check that *Exec satisfies the Git interface.
var _ Git = (*Exec)(nil)
