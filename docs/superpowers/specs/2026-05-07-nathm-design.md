# `nathm` — Git Branch Organization & Cleanup CLI

**Date:** 2026-05-07
**Status:** Design approved, pending implementation plan

## Summary

`nathm` (Arabic نظم — "to organize") is a TUI-first CLI tool for organizing and cleaning up local git branches. The default invocation opens an interactive list of all local branches with per-row actions (delete, force-delete, rename, checkout) and multi-select bulk delete. Subcommands provide scriptable entry points for common operations, especially pruning stale branches whose upstreams have been deleted or whose tips are merged into the base branch.

## Goals

- Make local branch cleanup fast and visible in repos that have accumulated many branches.
- One-command pruning of branches that are "done" — gone-tracking and/or merged into base.
- Safe by default: protected branches, confirmation before destructive ops, never reaches beyond the local repo.
- Scriptable for non-interactive use via subcommands and flags.

## Non-Goals (v1)

- Remote branch operations beyond what's needed for status (no `git push`, no remote rename, no PR creation).
- GitHub PR awareness (deferred to v2 — likely via `gh` shell-out).
- Cross-repo or worktree-aware operations.
- Editing config from within the tool.

## Tech Stack

- **Language:** Go (single static binary, easy distribution via `brew install` or `go install`).
- **TUI:** [Bubble Tea](https://github.com/charmbracelet/bubbletea) + Lip Gloss for styling. Same stack as `lazygit`, `gh`, and `glow`.
- **CLI framework:** [Cobra](https://github.com/spf13/cobra) for subcommand routing.
- **Git:** Shells out to `git` on `PATH`. We do *not* import `go-git`. Reasons: respects user's `gitconfig` and credential helpers, more reliable, gives us a natural test seam.
- **Config:** TOML, parsed via `BurntSushi/toml`.
- **Tests:** standard `testing` + Bubble Tea's `teatest` harness for TUI smoke tests.

## Architecture

### Layout

```
nathm/
├── main.go                     # entry point: cmd.Execute()
├── cmd/                        # Cobra command tree
│   ├── root.go                 # `nathm` (default → opens TUI)
│   ├── prune.go                # `nathm prune [--yes]`
│   ├── rename.go               # `nathm rename <old> <new>`
│   ├── list.go                 # `nathm list [--stale|--all]`
│   └── version.go
├── internal/
│   ├── git/                    # thin git CLI wrapper (interface + real impl)
│   ├── branch/                 # domain types + classification
│   ├── tui/                    # Bubble Tea models, views, keymaps
│   └── config/                 # load/save ~/.config/nathm/config.toml
├── docs/
└── go.mod
```

### Why this layout

- Each `internal` package has one job and one obvious test target.
- `internal/git` is the only place we shell out — and the only place we mock for tests.
- `internal/branch` knows nothing about TUI or shell.
- `internal/tui` knows nothing about `os/exec`.
- `cmd` is pure wiring — no business logic.

## Domain Model (`internal/branch`)

```go
type Branch struct {
    Name              string
    IsCurrent         bool
    Upstream          string    // "origin/feature-x" or "" if none
    UpstreamGone      bool      // tracking branch deleted upstream
    Ahead             int       // commits ahead of upstream (or base if no upstream)
    Behind            int
    LastCommitTime    time.Time
    LastCommitSHA     string
    LastCommitSubject string
    MergedIntoBase    bool      // tip reachable from main/master
    Protected         bool      // matches a protected pattern
}

type Status int
const (
    StatusActive  Status = iota // has work, not gone, not merged
    StatusGone                  // upstream deleted (PR merged & branch deleted on remote)
    StatusMerged                // tip reachable from base, but upstream not gone
    StatusBoth                  // gone AND merged
)

func (b Branch) Status() Status // derives from UpstreamGone + MergedIntoBase
```

### Population pipeline

1. `git for-each-ref --format='...' refs/heads` — single call gets every branch with name, upstream, upstream-gone bit, last-commit time/sha/subject.
2. Detect base branch: first match from `[main, master]` (default; user-overridable via `base_branches` config) that exists locally.
3. For each branch, `git rev-list --left-right --count base...branch` for ahead/behind. Batched if perf becomes an issue (premature for v1).
4. Apply protected matching: current branch + base names + user's config glob patterns.

The status enum drives row colors in the TUI and filter behavior in `prune`. The `Protected` flag gates destructive actions.

## TUI (`internal/tui`)

### Screens

```
┌──────────────┐  enter / a       ┌────────────────┐
│  List screen │ ───────────────▶ │  Action menu   │
│  (default)   │ ◀─── esc ────── │  (modal popup) │
└──────┬───────┘                  └────────┬───────┘
       │                                   │
       │ /                                 │ delete / force-delete
       ▼                                   ▼
┌──────────────┐                  ┌────────────────┐
│ Filter input │                  │ Confirm dialog │
│ (in-place)   │                  │ (y/N)          │
└──────────────┘                  └────────────────┘
                                           │
                                  rename ──┤
                                           ▼
                                  ┌────────────────┐
                                  │ Rename input   │
                                  │ (text field)   │
                                  └────────────────┘
```

### List screen

Columns:

| Col | Content |
|---|---|
| `[ ]` | selection checkbox |
| name | branch name; current branch shown as `* name` and dimmed |
| status | `gone` / `merged` / `gone+merged` / `active` / `protected` (color-coded) |
| age | relative, e.g. `3w` |
| ↑ ↓ | ahead/behind base |
| subject | last-commit subject, truncated to remaining width |

**Default sort:** stale-first (gone+merged > gone > merged > active), then by last-commit age descending. The cleanup use case lands the most-likely-to-delete branches at the top.

**Default filter:** none (show all).

### Keybindings

| Key | Action |
|---|---|
| `↑`/`↓` or `j`/`k` | navigate |
| `space` | toggle selection |
| `a` | select all visible (post-filter) |
| `A` | clear selection |
| `enter` | open action menu for cursor row |
| `d` | safe delete (cursor row, or all selected if any) |
| `D` | force delete |
| `r` | rename (cursor row only — never bulk) |
| `c` | checkout |
| `/` | filter |
| `s` | cycle sort (stale-first → name → age) |
| `p` | toggle "show only stale" filter |
| `?` | help overlay |
| `q` / `esc` | quit |

### Confirm dialog

Lists exactly what will happen and any blockers:

```
Delete 3 branches:
  feat/login-form    (gone, 3w old)
  fix/typo           (merged, 1mo old)
  ⚠ release/v1.2    skipped — protected pattern

Continue? [y/N]
```

### Refresh & error surfacing

- After every destructive op, the model re-runs the loaders and rebuilds the list.
- Per-branch failures in a batch don't abort the batch — they're collected and shown in a status bar.
- TUI panics use a deferred recover that restores terminal state before printing.

## Subcommands

```
nathm                          # TUI (default)
nathm prune [--yes]            # prune stale branches
nathm rename <old> <new>       # local rename
nathm list [--stale|--all]     # plain stdout, scriptable (TSV)
nathm version
```

### `nathm prune`

1. `git fetch --prune` (so "gone" tracking is current).
2. Compute candidates = `gone ∪ merged`, minus protected.
3. Print the list with reasons:
   ```
   The following 4 branches will be deleted:
     feat/login-form        gone           last commit 3w ago
     fix/typo-readme        merged         last commit 1mo ago
     experiment/grpc        gone+merged    last commit 2mo ago
     chore/bump-deps        merged         last commit 5d ago
   Continue? [y/N]
   ```
4. With `--yes`: skip the prompt.
5. Per-branch failures print to stderr and are skipped. Exit 0 if at least one succeeded; exit 1 only on total failure.

### `nathm rename`

- Local rename only in v1 via `git branch -m`.
- Refuses if the target name already exists locally.
- Remote rename is deferred. When added, it'll be a `--remote` flag (create new remote ref, delete old, update upstream).

### `nathm list`

- TSV output, one branch per line: `name\tstatus\tage_seconds\tahead\tbehind`.
- `--stale` filters to gone+merged.
- Useful for piping into other tools and as a smoke test of the data layer without touching the TUI.

### Exit codes

| Code | Meaning |
|---|---|
| 0 | success |
| 1 | runtime error (git missing, not a git repo, total failure) |
| 2 | user cancelled |

## Configuration

File: `${XDG_CONFIG_HOME:-~/.config}/nathm/config.toml`. All fields optional.

```toml
# Glob patterns matched per branch name. These branches cannot be deleted/renamed.
protected_patterns = ["release/*", "hotfix/*"]

# Override default base-branch detection order. First match that exists locally wins.
base_branches = ["main", "master"]

# stale-first | name | age
default_sort = "stale-first"
```

Auto-created with comments on first run if missing. No in-tool editor in v1.

### Default protected branches

Even with no config, the following are always protected:

- The current/checked-out branch.
- The detected base branch (`main` or `master`, whichever exists).

Users add more via `protected_patterns`.

## Error Handling

Three failure modes, three responses:

1. **Pre-flight failures** (git not on PATH, not a git repo, unreadable config): clear error to stderr, exit 1, never open the TUI.
2. **Per-branch failures in a batch**: collect, continue, surface in the status bar (TUI) or stderr (non-interactive).
3. **TUI-internal panics**: deferred recover restores terminal state before re-raising.

### Safety invariants

- Never call `git` with `--no-verify`.
- Never force-push.
- Never `git push` at all in v1.
- Never run a destructive op against a protected branch — protection check happens both in the UI (disabled rows) and again at the action layer.

## Testing Strategy

Three layers:

1. **Unit tests** (`internal/branch`): classification, protection matching, status derivation. Pure functions, no I/O.
2. **Integration tests** (`internal/git`): a `testGitRepo()` helper creates a tempdir repo, makes commits, sets up branches in known states (gone, merged, ahead, etc.). Real `git` invocations. Covers the brittle parsing logic.
3. **TUI smoke tests** (`internal/tui`): drive keystrokes via Bubble Tea's `teatest`, assert on rendered frames. Smoke-level only — exhaustive coverage is too brittle to maintain.

The shell-out interface in `internal/git` is also the seam tests can use a fake implementation against if a particular case is hard to set up with a real repo.

## Distribution

- `go install github.com/<user>/nathm@latest` — works for any Go user.
- Homebrew formula in a tap (deferred until we have a release worth shipping).
- GitHub Releases with prebuilt binaries for darwin/amd64, darwin/arm64, linux/amd64, linux/arm64 (deferred — start with `go install`).

## Deferred to v2+

- GitHub PR awareness (PR number, state, mergeability via `gh` shell-out).
- Remote rename (`nathm rename --remote`).
- Worktree-aware ops.
- In-tool config editor.
- Integration with `git stash` / `git reflog` for "undo last delete".
- Branch grouping by prefix (e.g., collapsing all `feat/*` under one heading).
