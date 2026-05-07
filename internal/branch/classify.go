package branch

import "path"

// DetectBase returns the first preferred name that exists in `all`.
// Returns "" if none match. `all` is the full list of local branch names.
func DetectBase(preferred, all []string) string {
	have := make(map[string]struct{}, len(all))
	for _, n := range all {
		have[n] = struct{}{}
	}
	for _, p := range preferred {
		if _, ok := have[p]; ok {
			return p
		}
	}
	return ""
}

// MarkProtected sets b.Protected = true on any branch matching:
//   - the current branch
//   - the base branch name
//   - any glob in patterns (using path.Match — supports *, ?, [class])
func MarkProtected(branches []Branch, base string, patterns []string) {
	for i := range branches {
		b := &branches[i]
		if b.IsCurrent {
			b.Protected = true
			continue
		}
		if base != "" && b.Name == base {
			b.Protected = true
			continue
		}
		for _, p := range patterns {
			if matched, _ := path.Match(p, b.Name); matched {
				b.Protected = true
				break
			}
		}
	}
}
