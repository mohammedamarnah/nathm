package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Point XDG_CONFIG_HOME at an empty tempdir → no config file.
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir) // safety: prevent fallback to real home

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	want := Config{
		BaseBranches:      []string{"main", "master"},
		ProtectedPatterns: []string{},
		DefaultSort:       "stale-first",
	}
	if !reflect.DeepEqual(cfg, want) {
		t.Fatalf("got %+v, want %+v", cfg, want)
	}
	// Also check the file was created.
	if _, err := os.Stat(filepath.Join(dir, "nathm", "config.toml")); err != nil {
		t.Fatalf("expected default config to be created: %v", err)
	}
}

func TestLoad_OverridesFromFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)

	cfgDir := filepath.Join(dir, "nathm")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	contents := `protected_patterns = ["release/*", "hotfix/*"]
base_branches = ["trunk"]
default_sort = "name"
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg.ProtectedPatterns, []string{"release/*", "hotfix/*"}) {
		t.Errorf("patterns = %v", cfg.ProtectedPatterns)
	}
	if !reflect.DeepEqual(cfg.BaseBranches, []string{"trunk"}) {
		t.Errorf("bases = %v", cfg.BaseBranches)
	}
	if cfg.DefaultSort != "name" {
		t.Errorf("sort = %q", cfg.DefaultSort)
	}
}
