package git

import (
	"os/exec"
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

// Compile-time check that *Exec satisfies the Git interface.
var _ Git = (*Exec)(nil)
