// Package branch defines the domain model for git branches as nathm sees them.
//
// This package is pure: no shell-outs, no I/O. It is the input to TUI rendering
// and the output of the loader in load.go.
package branch

import "time"

// Status classifies a branch for cleanup decisions.
type Status int

const (
	StatusActive Status = iota // has work, not gone, not merged
	StatusGone                 // upstream tracking branch deleted
	StatusMerged               // tip reachable from base, but upstream not gone
	StatusBoth                 // gone AND merged
)

func (s Status) String() string {
	switch s {
	case StatusGone:
		return "gone"
	case StatusMerged:
		return "merged"
	case StatusBoth:
		return "gone+merged"
	default:
		return "active"
	}
}

// Branch is the in-memory representation of a single local branch.
type Branch struct {
	Name              string
	IsCurrent         bool
	Upstream          string // "origin/foo" or "" if none
	UpstreamGone      bool   // tracking branch deleted upstream
	Ahead             int    // ahead of upstream, or base if no upstream
	Behind            int
	LastCommitTime    time.Time
	LastCommitSHA     string
	LastCommitSubject string
	MergedIntoBase    bool
	Protected         bool
}

func (b Branch) Status() Status {
	switch {
	case b.UpstreamGone && b.MergedIntoBase:
		return StatusBoth
	case b.UpstreamGone:
		return StatusGone
	case b.MergedIntoBase:
		return StatusMerged
	default:
		return StatusActive
	}
}

// IsStale reports whether this branch is a candidate for prune cleanup.
func (b Branch) IsStale() bool {
	return b.UpstreamGone || b.MergedIntoBase
}
