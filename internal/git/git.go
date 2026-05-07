// Package git provides a thin wrapper over the git CLI.
//
// We deliberately shell out to git rather than importing go-git. Reasons:
//   - respects the user's gitconfig and credential helpers
//   - more reliable for niche behavior (e.g. upstream:track)
//   - gives us a natural test seam: real impl in tests, fake impl in unit tests
package git

import "github.com/USER/nathm/internal/branch"

// Git is the surface area nathm needs from the git CLI.
// Methods are added as features need them.
type Git interface {
	IsRepo() bool
	ListBranches() ([]branch.Branch, error)
	AheadBehind(branch, base string) (ahead, behind int, err error)
	MergedInto(branch, base string) (bool, error)
}
