package branch

import (
	"testing"
)

func TestDetectBase(t *testing.T) {
	tests := []struct {
		name      string
		preferred []string
		all       []string
		want      string
	}{
		{"main wins over master", []string{"main", "master"}, []string{"feature", "master", "main"}, "main"},
		{"falls back to master", []string{"main", "master"}, []string{"feature", "master"}, "master"},
		{"none found", []string{"main", "master"}, []string{"feature"}, ""},
		{"custom order respected", []string{"trunk", "main"}, []string{"main", "trunk"}, "trunk"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectBase(tt.preferred, tt.all); got != tt.want {
				t.Fatalf("DetectBase = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMarkProtected(t *testing.T) {
	bs := []Branch{
		{Name: "main", IsCurrent: true},
		{Name: "master"},
		{Name: "feature/foo"},
		{Name: "release/v1.0"},
		{Name: "hotfix/abc"},
		{Name: "wip", IsCurrent: false},
	}
	MarkProtected(bs, "main", []string{"release/*", "hotfix/*"})

	got := map[string]bool{}
	for _, b := range bs {
		got[b.Name] = b.Protected
	}
	want := map[string]bool{
		"main":         true, // current AND base
		"master":       false,
		"feature/foo":  false,
		"release/v1.0": true,
		"hotfix/abc":   true,
		"wip":          false,
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("Protected[%s] = %v, want %v", k, got[k], v)
		}
	}
}

func TestMarkProtected_BaseAlwaysProtected(t *testing.T) {
	bs := []Branch{{Name: "main"}, {Name: "feature"}}
	MarkProtected(bs, "main", nil)
	if !bs[0].Protected {
		t.Fatalf("base branch should be protected")
	}
	if bs[1].Protected {
		t.Fatalf("non-base, non-current should not be protected")
	}
}
