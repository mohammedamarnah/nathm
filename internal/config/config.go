// Package config loads nathm's user configuration from TOML.
//
// The file lives at ${XDG_CONFIG_HOME:-$HOME/.config}/nathm/config.toml.
// Missing fields fall back to defaults. Missing file is auto-created with
// commented defaults on first run.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	BaseBranches      []string `toml:"base_branches"`
	ProtectedPatterns []string `toml:"protected_patterns"`
	DefaultSort       string   `toml:"default_sort"`
}

const defaultTOML = `# nathm configuration
# https://github.com/USER/nathm

# Glob patterns for branches that should never be deleted or renamed.
# Examples: "release/*", "hotfix/*", "user/<your-name>/*"
protected_patterns = []

# Order of preference for detecting the repo's base branch. The first match
# that exists locally is used for ahead/behind and merge-status computation,
# and is always treated as protected.
base_branches = ["main", "master"]

# Default sort in the TUI: "stale-first" | "name" | "age"
default_sort = "stale-first"
`

func defaults() Config {
	return Config{
		BaseBranches:      []string{"main", "master"},
		ProtectedPatterns: []string{},
		DefaultSort:       "stale-first",
	}
}

// Path returns the absolute path to the config file, honoring XDG.
func Path() (string, error) {
	root := os.Getenv("XDG_CONFIG_HOME")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		root = filepath.Join(home, ".config")
	}
	return filepath.Join(root, "nathm", "config.toml"), nil
}

// Load reads config, applying defaults for missing fields. If the file does
// not exist, it is created with default contents.
func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Config{}, err
	}
	cfg := defaults()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		if err := writeDefault(path); err != nil {
			return Config{}, fmt.Errorf("write default config: %w", err)
		}
		return cfg, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if cfg.ProtectedPatterns == nil {
		cfg.ProtectedPatterns = []string{}
	}
	if len(cfg.BaseBranches) == 0 {
		cfg.BaseBranches = defaults().BaseBranches
	}
	if cfg.DefaultSort == "" {
		cfg.DefaultSort = defaults().DefaultSort
	}
	return cfg, nil
}

func writeDefault(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(defaultTOML), 0o644)
}
