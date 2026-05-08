# nathm

`nathm` (Arabic نَظْم — *to organize*) is a TUI for organization and management of git branches.

![demo](docs/demo.gif)

## Install

Make sure that Go is installed and is in the `PATH`:
```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
```

```bash
go install github.com/mohammedamarnah/nathm@latest
```

## Usage

### Interactive

In any git repo:

```bash
nathm
```

Keys:

| Key | Action |
|---|---|
| `↑`/`↓` or `j`/`k` | navigate |
| `space` | toggle selection |
| `a` | select all visible |
| `A` | clear selection |
| `d` | safe delete |
| `D` | force delete |
| `r` | rename |
| `c` | checkout |
| `/` | filter |
| `s` | cycle sort |
| `p` | toggle stale-only |
| `?` | help overlay |
| `q` | quit |

### Subcommands

```bash
nathm prune           # confirm-before-delete cleanup of stale branches
nathm prune --yes     # skip confirmation
nathm rename old new  # rename a local branch
nathm list            # TSV output (name, status, age_seconds, ahead, behind)
nathm list --stale    # only stale branches
nathm version
```

## Configuration

`${XDG_CONFIG_HOME:-$HOME/.config}/nathm/config.toml`. Auto-created on first run. Fields:

- `protected_patterns` — globs for branches that can never be deleted/renamed.
- `base_branches` — preference order for base branch detection.
- `default_sort` — `stale-first` | `name` | `age`.

