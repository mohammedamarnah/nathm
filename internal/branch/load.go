package branch

// Gitter is the slice of git operations the loader needs. It mirrors a subset
// of git.Git but lives here to avoid an import cycle.
type Gitter interface {
	ListBranches() ([]Branch, error)
	AheadBehind(branch, base string) (ahead, behind int, err error)
	MergedInto(branch, base string) (bool, error)
}

// LoadConfig is the input config for Load. Field names match config.Config so
// callers can pass the value through.
type LoadConfig struct {
	BaseBranches      []string // ordered preference, e.g. ["main", "master"]
	ProtectedPatterns []string // glob patterns
}

// Load returns the fully-classified branch list for the repo.
func Load(g Gitter, cfg LoadConfig) ([]Branch, error) {
	bs, err := g.ListBranches()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(bs))
	for i, b := range bs {
		names[i] = b.Name
	}
	base := DetectBase(cfg.BaseBranches, names)

	for i := range bs {
		b := &bs[i]
		// Backfill ahead/behind for branches without upstream tracking.
		if b.Upstream == "" && base != "" && b.Name != base {
			a, beh, err := g.AheadBehind(b.Name, base)
			if err == nil {
				b.Ahead = a
				b.Behind = beh
			}
		}
		// Compute merged-into-base.
		if base != "" && b.Name != base {
			merged, err := g.MergedInto(b.Name, base)
			if err == nil && merged {
				b.MergedIntoBase = true
			}
		}
	}

	MarkProtected(bs, base, cfg.ProtectedPatterns)
	return bs, nil
}
