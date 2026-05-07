package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/USER/nathm/internal/branch"
)

func TestModel_View_RendersBranches(t *testing.T) {
	bs := []branch.Branch{
		{Name: "main", IsCurrent: true, LastCommitTime: time.Now()},
		{Name: "feature/foo", LastCommitTime: time.Now()},
		{Name: "stale", UpstreamGone: true, LastCommitTime: time.Now()},
	}
	m := NewModel(bs, nil) // nil git (we don't act yet)
	m.SetSize(120, 30)
	out := m.View()
	for _, name := range []string{"main", "feature/foo", "stale"} {
		if !strings.Contains(out, name) {
			t.Errorf("view missing %q:\n%s", name, out)
		}
	}
}
