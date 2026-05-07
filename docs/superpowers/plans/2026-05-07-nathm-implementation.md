# nathm Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `nathm`, a TUI-first CLI for organizing and cleaning up local git branches, with subcommands for scriptable use.

**Architecture:** Go binary that shells out to `git`. Cobra command tree at `cmd/`. Domain logic in `internal/branch`. Git shell-out wrapper with an interface seam for testing in `internal/git`. Bubble Tea TUI in `internal/tui`. TOML config in `internal/config`. The `internal/git` package is the only place we shell out and the only place we mock for tests.

**Tech Stack:** Go 1.22+, Cobra (CLI), Bubble Tea + Bubbles + Lip Gloss (TUI), BurntSushi/toml (config), dustin/go-humanize (time formatting). Module path uses placeholder `github.com/USER/nathm` — engineer should replace `USER` with the actual GitHub user/org name before publishing.

**Spec:** `docs/superpowers/specs/2026-05-07-nathm-design.md`

---

## File Structure

```
nathm/
├── go.mod
├── go.sum
├── main.go                              # tiny entry point
├── cmd/
│   ├── root.go                          # root command (default → TUI)
│   ├── version.go                       # `nathm version`
│   ├── list.go                          # `nathm list`
│   ├── prune.go                         # `nathm prune`
│   └── rename.go                        # `nathm rename`
├── internal/
│   ├── git/
│   │   ├── git.go                       # Git interface
│   │   ├── exec.go                      # real shell-out impl
│   │   ├── testhelp_test.go             # tempdir repo helpers
│   │   └── exec_test.go                 # integration tests
│   ├── branch/
│   │   ├── branch.go                    # Branch struct, Status enum
│   │   ├── parse.go                     # parse for-each-ref output
│   │   ├── classify.go                  # base detection, protection, status derivation
│   │   ├── load.go                      # orchestrator: Load(git, cfg) → []Branch
│   │   └── *_test.go                    # tests per file
│   ├── config/
│   │   ├── config.go                    # Config struct + Load()
│   │   └── config_test.go
│   └── tui/
│       ├── model.go                     # main Model, Init, Update, View
│       ├── keymap.go                    # key bindings
│       ├── messages.go                  # tea.Msg types
│       ├── styles.go                    # lipgloss styles
│       └── model_test.go                # teatest smoke tests
└── docs/
    └── superpowers/
        ├── specs/2026-05-07-nathm-design.md
        └── plans/2026-05-07-nathm-implementation.md
```

---

## Conventions

- **Module path:** `github.com/USER/nathm` (placeholder — replace before publish).
- **Commit style:** Conventional commits: `feat:`, `test:`, `refactor:`, `docs:`.
- **Test naming:** `TestThing_Scenario_Expected`.
- **Always `git add` specific paths**, never `git add -A` or `git add .`, to avoid committing junk.
- **Run `go test ./...`** after each task before committing.
- **TDD:** every task starts with a failing test.

---

## Task 1: Project bootstrap + version subcommand

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `cmd/root.go`
- Create: `cmd/version.go`

- [ ] **Step 1: Initialize the module and add Cobra**

```bash
cd /Users/mohammadalamarneh/workspace/nathm
go mod init github.com/USER/nathm
go get github.com/spf13/cobra@latest
```

- [ ] **Step 2: Create main.go**

```go
package main

import "github.com/USER/nathm/cmd"

func main() {
	cmd.Execute()
}
```

- [ ] **Step 3: Create cmd/root.go**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nathm",
	Short: "Organize and clean up local git branches",
	Long:  "nathm (نظم) is an interactive TUI and CLI for organizing local git branches.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TUI wired up in Task 23. For now, print a placeholder.
		fmt.Fprintln(cmd.OutOrStdout(), "TUI not yet wired — try `nathm version` or `nathm --help`.")
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Create cmd/version.go with version constant**

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "0.1.0-dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version and exit",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), "nathm", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
```

- [ ] **Step 5: Verify build and version output**

Run: `go build ./... && ./nathm version`
Expected: `nathm 0.1.0-dev`

Run: `./nathm --help`
Expected: help text listing `version` subcommand.

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum main.go cmd/root.go cmd/version.go
git commit -m "feat: bootstrap nathm binary with cobra and version subcommand"
```

---

## Task 2: Define the Git interface

**Files:**
- Create: `internal/git/git.go`

The Git interface is the seam that lets us test the rest of the system without shelling out. We start with one method and grow it.

- [ ] **Step 1: Create internal/git/git.go**

```go
// Package git provides a thin wrapper over the git CLI.
//
// We deliberately shell out to git rather than importing go-git. Reasons:
//   - respects the user's gitconfig and credential helpers
//   - more reliable for niche behavior (e.g. upstream:track)
//   - gives us a natural test seam: real impl in tests, fake impl in unit tests
package git

// Git is the surface area nathm needs from the git CLI.
// Methods are added as features need them.
type Git interface {
	IsRepo() bool
}
```

- [ ] **Step 2: Verify the package compiles**

Run: `go build ./internal/git/`
Expected: no output, exit 0.

- [ ] **Step 3: Commit**

```bash
git add internal/git/git.go
git commit -m "feat(git): introduce Git interface seam"
```

---

## Task 3: Implement Git.IsRepo + tempdir test helper

**Files:**
- Create: `internal/git/exec.go`
- Create: `internal/git/testhelp_test.go`
- Create: `internal/git/exec_test.go`

This task establishes the integration test pattern (real `git` against a tempdir repo) we'll use throughout.

- [ ] **Step 1: Write the failing test for IsRepo**

Create `internal/git/exec_test.go`:

```go
package git

import (
	"testing"
)

func TestExec_IsRepo_TrueInsideRepo(t *testing.T) {
	dir := newTestRepo(t)
	g := NewExec(dir)
	if !g.IsRepo() {
		t.Fatalf("expected IsRepo() = true inside %s", dir)
	}
}

func TestExec_IsRepo_FalseOutsideRepo(t *testing.T) {
	dir := t.TempDir() // not initialized
	g := NewExec(dir)
	if g.IsRepo() {
		t.Fatalf("expected IsRepo() = false in non-repo dir")
	}
}
```

- [ ] **Step 2: Write the test helper**

Create `internal/git/testhelp_test.go`:

```go
package git

import (
	"os/exec"
	"testing"
)

// newTestRepo creates a fresh git repo in a tempdir with one commit on `main`.
// Caller gets the directory path. Cleanup is automatic via t.TempDir().
func newTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runIn(t, dir, "git", "init", "-q", "-b", "main")
	runIn(t, dir, "git", "config", "user.email", "nathm-test@example.com")
	runIn(t, dir, "git", "config", "user.name", "nathm test")
	runIn(t, dir, "git", "config", "commit.gpgsign", "false")
	runIn(t, dir, "git", "commit", "--allow-empty", "-q", "-m", "init")
	return dir
}

func runIn(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v in %s: %v\n%s", name, args, dir, err, out)
	}
}
```

- [ ] **Step 3: Run the test and verify it fails**

Run: `go test ./internal/git/ -run TestExec_IsRepo`
Expected: FAIL — `NewExec` and `IsRepo` not defined.

- [ ] **Step 4: Implement Exec.IsRepo**

Create `internal/git/exec.go`:

```go
package git

import (
	"os/exec"
)

// Exec is the real git CLI wrapper.
type Exec struct {
	dir string // working directory; "" means the process cwd
}

// NewExec returns an Exec wrapper rooted at dir. Pass "" to use the current
// working directory.
func NewExec(dir string) *Exec {
	return &Exec{dir: dir}
}

func (e *Exec) IsRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = e.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return string(out) == "true\n"
}
```

- [ ] **Step 5: Update the Git interface to include IsRepo (already there) and add `*Exec` as an implementer**

Already done in Task 2 — interface has `IsRepo()`. Verify `*Exec` satisfies the interface by adding a compile-time check:

Append to `internal/git/exec.go`:

```go
var _ Git = (*Exec)(nil)
```

- [ ] **Step 6: Run tests and verify pass**

Run: `go test ./internal/git/ -v`
Expected: both `TestExec_IsRepo_*` PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/git/exec.go internal/git/exec_test.go internal/git/testhelp_test.go
git commit -m "feat(git): implement Exec.IsRepo with tempdir test helper"
```

---

## Task 4: Branch and Status types

**Files:**
- Create: `internal/branch/branch.go`
- Create: `internal/branch/branch_test.go`

- [ ] **Step 1: Write the failing test for Status derivation**

Create `internal/branch/branch_test.go`:

```go
package branch

import "testing"

func TestBranch_Status(t *testing.T) {
	tests := []struct {
		name string
		b    Branch
		want Status
	}{
		{"active", Branch{}, StatusActive},
		{"gone only", Branch{UpstreamGone: true}, StatusGone},
		{"merged only", Branch{MergedIntoBase: true}, StatusMerged},
		{"both", Branch{UpstreamGone: true, MergedIntoBase: true}, StatusBoth},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.Status(); got != tt.want {
				t.Fatalf("Status() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatus_String(t *testing.T) {
	cases := map[Status]string{
		StatusActive: "active",
		StatusGone:   "gone",
		StatusMerged: "merged",
		StatusBoth:   "gone+merged",
	}
	for s, want := range cases {
		if got := s.String(); got != want {
			t.Fatalf("Status(%d).String() = %q, want %q", s, got, want)
		}
	}
}
```

- [ ] **Step 2: Run the test and verify it fails**

Run: `go test ./internal/branch/`
Expected: FAIL — `Branch`, `Status`, etc. undefined.

- [ ] **Step 3: Implement the types**

Create `internal/branch/branch.go`:

```go
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
```

- [ ] **Step 4: Run tests and verify pass**

Run: `go test ./internal/branch/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/branch/branch.go internal/branch/branch_test.go
git commit -m "feat(branch): add Branch and Status types with classification"
```

---

## Task 5: Parse `for-each-ref` output

**Files:**
- Create: `internal/branch/parse.go`
- Create: `internal/branch/parse_test.go`

Pure parser, no git shell-out. Easy to test deterministically with fixture strings. We use NUL (`\x00`) as field separator and `\n` between records.

The format string we'll feed to git later:

```
%(HEAD) %00 %(refname:short) %00 %(upstream:short) %00 %(upstream:track) %00 %(committerdate:unix) %00 %(objectname) %00 %(contents:subject)
```

(`%(HEAD)` is `*` for the current branch and a space otherwise. `%(upstream:track)` is `[gone]`, `[ahead N]`, `[behind N]`, `[ahead N, behind M]`, `[]`, or empty.)

- [ ] **Step 1: Write failing tests**

Create `internal/branch/parse_test.go`:

```go
package branch

import (
	"testing"
	"time"
)

func TestParseForEachRef_Active(t *testing.T) {
	// fields: HEAD ref upstream track date sha subject
	in := []byte("*\x00main\x00origin/main\x00[ahead 2]\x001700000000\x00abc123\x00init\n")
	got, err := ParseForEachRef(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 branch, got %d", len(got))
	}
	b := got[0]
	if b.Name != "main" || !b.IsCurrent || b.Upstream != "origin/main" {
		t.Fatalf("unexpected branch: %+v", b)
	}
	if b.Ahead != 2 || b.Behind != 0 || b.UpstreamGone {
		t.Fatalf("ahead/behind/gone wrong: %+v", b)
	}
	if !b.LastCommitTime.Equal(time.Unix(1700000000, 0)) {
		t.Fatalf("time wrong: %v", b.LastCommitTime)
	}
	if b.LastCommitSHA != "abc123" || b.LastCommitSubject != "init" {
		t.Fatalf("sha/subject wrong: %+v", b)
	}
}

func TestParseForEachRef_Gone(t *testing.T) {
	in := []byte(" \x00feature\x00origin/feature\x00[gone]\x001700000000\x00def456\x00wip\n")
	got, _ := ParseForEachRef(in)
	if !got[0].UpstreamGone {
		t.Fatalf("want UpstreamGone, got %+v", got[0])
	}
}

func TestParseForEachRef_NoUpstream(t *testing.T) {
	in := []byte(" \x00local-only\x00\x00\x001700000000\x00aaa\x00msg\n")
	got, _ := ParseForEachRef(in)
	b := got[0]
	if b.Upstream != "" || b.UpstreamGone || b.Ahead != 0 || b.Behind != 0 {
		t.Fatalf("expected zero-valued upstream fields, got %+v", b)
	}
}

func TestParseForEachRef_AheadBehindBoth(t *testing.T) {
	in := []byte(" \x00x\x00origin/x\x00[ahead 3, behind 5]\x001700000000\x00aaa\x00msg\n")
	got, _ := ParseForEachRef(in)
	if got[0].Ahead != 3 || got[0].Behind != 5 {
		t.Fatalf("ahead/behind = %d/%d, want 3/5", got[0].Ahead, got[0].Behind)
	}
}

func TestParseForEachRef_MultipleBranches(t *testing.T) {
	in := []byte("*\x00main\x00\x00\x001700000000\x00aaa\x00msg\n \x00other\x00\x00\x001700000001\x00bbb\x00other msg\n")
	got, _ := ParseForEachRef(in)
	if len(got) != 2 {
		t.Fatalf("want 2 branches, got %d", len(got))
	}
}

func TestParseForEachRef_SubjectWithSpaces(t *testing.T) {
	in := []byte(" \x00x\x00\x00\x001700000000\x00aaa\x00fix: handle empty input\n")
	got, _ := ParseForEachRef(in)
	if got[0].LastCommitSubject != "fix: handle empty input" {
		t.Fatalf("subject = %q", got[0].LastCommitSubject)
	}
}
```

- [ ] **Step 2: Run tests and verify failure**

Run: `go test ./internal/branch/ -run ParseForEachRef`
Expected: FAIL — `ParseForEachRef` undefined.

- [ ] **Step 3: Implement parser**

Create `internal/branch/parse.go`:

```go
package branch

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// ParseForEachRef parses output from:
//
//	git for-each-ref \
//	  --format='%(HEAD)%00%(refname:short)%00%(upstream:short)%00%(upstream:track)%00%(committerdate:unix)%00%(objectname)%00%(contents:subject)' \
//	  refs/heads
//
// Each record is one line with NUL-separated fields. We split on `\n` then `\x00`.
func ParseForEachRef(out []byte) ([]Branch, error) {
	out = bytes.TrimRight(out, "\n")
	if len(out) == 0 {
		return nil, nil
	}
	lines := bytes.Split(out, []byte{'\n'})
	branches := make([]Branch, 0, len(lines))
	for i, line := range lines {
		fields := bytes.Split(line, []byte{0})
		if len(fields) != 7 {
			return nil, fmt.Errorf("line %d: expected 7 fields, got %d (%q)", i+1, len(fields), line)
		}
		b := Branch{
			Name:              string(fields[1]),
			IsCurrent:         string(fields[0]) == "*",
			Upstream:          string(fields[2]),
			LastCommitSHA:     string(fields[5]),
			LastCommitSubject: string(fields[6]),
		}
		track := string(fields[3])
		gone, ahead, behind, err := parseTrack(track)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		b.UpstreamGone = gone
		b.Ahead = ahead
		b.Behind = behind
		if ts := string(fields[4]); ts != "" {
			n, err := strconv.ParseInt(ts, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: bad date %q: %w", i+1, ts, err)
			}
			b.LastCommitTime = time.Unix(n, 0)
		}
		branches = append(branches, b)
	}
	return branches, nil
}

var (
	reAhead  = regexp.MustCompile(`ahead (\d+)`)
	reBehind = regexp.MustCompile(`behind (\d+)`)
)

// parseTrack interprets the upstream:track format.
//
// Possible values:
//   ""                          → no upstream
//   "[]"                        → up to date
//   "[gone]"                    → upstream deleted
//   "[ahead N]"
//   "[behind N]"
//   "[ahead N, behind M]"
func parseTrack(t string) (gone bool, ahead, behind int, err error) {
	if t == "" {
		return false, 0, 0, nil
	}
	if t == "[gone]" {
		return true, 0, 0, nil
	}
	if m := reAhead.FindStringSubmatch(t); m != nil {
		ahead, _ = strconv.Atoi(m[1])
	}
	if m := reBehind.FindStringSubmatch(t); m != nil {
		behind, _ = strconv.Atoi(m[1])
	}
	return false, ahead, behind, nil
}
```

- [ ] **Step 4: Run tests and verify pass**

Run: `go test ./internal/branch/ -v`
Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/branch/parse.go internal/branch/parse_test.go
git commit -m "feat(branch): parse git for-each-ref output"
```

---

## Task 6: Add `ListBranches` to the Git interface

**Files:**
- Modify: `internal/git/git.go`
- Modify: `internal/git/exec.go`
- Modify: `internal/git/exec_test.go`

This connects `internal/git` to `internal/branch` for the first time. We expose `ListBranches() ([]branch.Branch, error)`.

- [ ] **Step 1: Write the failing test**

Append to `internal/git/exec_test.go`:

```go
import (
	// ... existing imports ...
	"sort"
)

func TestExec_ListBranches_OneBranch(t *testing.T) {
	dir := newTestRepo(t)
	g := NewExec(dir)
	bs, err := g.ListBranches()
	if err != nil {
		t.Fatal(err)
	}
	if len(bs) != 1 || bs[0].Name != "main" || !bs[0].IsCurrent {
		t.Fatalf("got %+v", bs)
	}
}

func TestExec_ListBranches_MultipleBranches(t *testing.T) {
	dir := newTestRepo(t)
	runIn(t, dir, "git", "branch", "feature-a")
	runIn(t, dir, "git", "branch", "feature-b")
	g := NewExec(dir)
	bs, err := g.ListBranches()
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, len(bs))
	for i, b := range bs {
		names[i] = b.Name
	}
	sort.Strings(names)
	if got, want := strings.Join(names, ","), "feature-a,feature-b,main"; got != want {
		t.Fatalf("names = %q, want %q", got, want)
	}
}
```

(Add `"strings"` to the imports if not already present.)

- [ ] **Step 2: Update the interface**

Edit `internal/git/git.go`:

```go
package git

import "github.com/USER/nathm/internal/branch"

type Git interface {
	IsRepo() bool
	ListBranches() ([]branch.Branch, error)
}
```

- [ ] **Step 3: Implement ListBranches**

Append to `internal/git/exec.go`:

```go
import (
	// existing imports + these:
	"github.com/USER/nathm/internal/branch"
)

const forEachRefFormat = "%(HEAD)%00%(refname:short)%00%(upstream:short)%00%(upstream:track)%00%(committerdate:unix)%00%(objectname)%00%(contents:subject)"

func (e *Exec) ListBranches() ([]branch.Branch, error) {
	cmd := exec.Command("git", "for-each-ref", "--format="+forEachRefFormat, "refs/heads")
	cmd.Dir = e.dir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git for-each-ref: %w", err)
	}
	return branch.ParseForEachRef(out)
}
```

(Add `"fmt"` to the imports.)

- [ ] **Step 4: Run tests and verify pass**

Run: `go test ./internal/git/ -v`
Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/git/git.go internal/git/exec.go internal/git/exec_test.go
git commit -m "feat(git): list local branches via for-each-ref"
```

---

## Task 7: Compute ahead/behind vs base

**Files:**
- Modify: `internal/git/git.go`
- Modify: `internal/git/exec.go`
- Modify: `internal/git/exec_test.go`

For branches without an upstream (or with `[gone]`), we need to compute ahead/behind vs the base branch. `git rev-list --left-right --count base...branch` gives `behind\tahead` separated by a tab.

- [ ] **Step 1: Write failing test**

Append to `internal/git/exec_test.go`:

```go
func TestExec_AheadBehind(t *testing.T) {
	dir := newTestRepo(t)
	// add 2 commits on main
	writeFile(t, dir, "a.txt", "a")
	runIn(t, dir, "git", "add", "a.txt")
	runIn(t, dir, "git", "commit", "-q", "-m", "a")
	// branch off and add 1 commit
	runIn(t, dir, "git", "checkout", "-q", "-b", "feature")
	writeFile(t, dir, "b.txt", "b")
	runIn(t, dir, "git", "add", "b.txt")
	runIn(t, dir, "git", "commit", "-q", "-m", "b")
	// add another commit on main
	runIn(t, dir, "git", "checkout", "-q", "main")
	writeFile(t, dir, "c.txt", "c")
	runIn(t, dir, "git", "add", "c.txt")
	runIn(t, dir, "git", "commit", "-q", "-m", "c")

	g := NewExec(dir)
	ahead, behind, err := g.AheadBehind("feature", "main")
	if err != nil {
		t.Fatal(err)
	}
	if ahead != 1 || behind != 1 {
		t.Fatalf("ahead/behind = %d/%d, want 1/1", ahead, behind)
	}
}
```

Add a `writeFile` helper in `internal/git/testhelp_test.go`:

```go
import (
	// existing imports +
	"os"
	"path/filepath"
)

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Run test and verify failure**

Run: `go test ./internal/git/ -run AheadBehind`
Expected: FAIL — method undefined.

- [ ] **Step 3: Add to interface and implement**

Edit `internal/git/git.go`:

```go
type Git interface {
	IsRepo() bool
	ListBranches() ([]branch.Branch, error)
	AheadBehind(branch, base string) (ahead, behind int, err error)
}
```

Append to `internal/git/exec.go`:

```go
import (
	// existing imports +
	"strings"
	"strconv"
)

func (e *Exec) AheadBehind(br, base string) (int, int, error) {
	cmd := exec.Command("git", "rev-list", "--left-right", "--count", base+"..."+br)
	cmd.Dir = e.dir
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("git rev-list: %w", err)
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected rev-list output: %q", out)
	}
	behind, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse behind: %w", err)
	}
	ahead, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse ahead: %w", err)
	}
	return ahead, behind, nil
}
```

- [ ] **Step 4: Run tests and verify pass**

Run: `go test ./internal/git/ -v`
Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/git/git.go internal/git/exec.go internal/git/exec_test.go internal/git/testhelp_test.go
git commit -m "feat(git): compute ahead/behind vs base branch"
```

---

## Task 8: Base detection, classification, and protection

**Files:**
- Create: `internal/branch/classify.go`
- Create: `internal/branch/classify_test.go`

Pure functions: `DetectBase`, `MarkProtected`, `MarkMerged`. No git calls — they take inputs already gathered.

- [ ] **Step 1: Write failing tests**

Create `internal/branch/classify_test.go`:

```go
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
```

- [ ] **Step 2: Run tests and verify failure**

Run: `go test ./internal/branch/ -run "DetectBase|MarkProtected"`
Expected: FAIL — functions undefined.

- [ ] **Step 3: Implement**

Create `internal/branch/classify.go`:

```go
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
```

- [ ] **Step 4: Run tests and verify pass**

Run: `go test ./internal/branch/ -v`
Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/branch/classify.go internal/branch/classify_test.go
git commit -m "feat(branch): detect base branch and mark protected branches"
```

---

## Task 9: Branch loader orchestrator

**Files:**
- Create: `internal/branch/load.go`
- Create: `internal/branch/load_test.go`
- Modify: `internal/git/git.go` (add `MergedInto` method)
- Modify: `internal/git/exec.go`
- Modify: `internal/git/exec_test.go`

The loader pulls everything together. We need one more git call: a way to test "is this branch merged into base?" — simplest is `git merge-base --is-ancestor <branch-tip> <base>`.

We also do ahead/behind backfill: for branches whose track string was empty (no upstream), compute vs base.

- [ ] **Step 1: Write failing test for MergedInto**

Append to `internal/git/exec_test.go`:

```go
func TestExec_MergedInto(t *testing.T) {
	dir := newTestRepo(t)
	runIn(t, dir, "git", "checkout", "-q", "-b", "feature")
	writeFile(t, dir, "f.txt", "f")
	runIn(t, dir, "git", "add", "f.txt")
	runIn(t, dir, "git", "commit", "-q", "-m", "f")
	runIn(t, dir, "git", "checkout", "-q", "main")
	runIn(t, dir, "git", "merge", "--no-ff", "-q", "-m", "merge feature", "feature")

	g := NewExec(dir)
	merged, err := g.MergedInto("feature", "main")
	if err != nil {
		t.Fatal(err)
	}
	if !merged {
		t.Fatal("feature should be merged into main")
	}

	// unmerged branch
	runIn(t, dir, "git", "checkout", "-q", "-b", "unmerged")
	writeFile(t, dir, "u.txt", "u")
	runIn(t, dir, "git", "add", "u.txt")
	runIn(t, dir, "git", "commit", "-q", "-m", "u")
	merged, err = g.MergedInto("unmerged", "main")
	if err != nil {
		t.Fatal(err)
	}
	if merged {
		t.Fatal("unmerged should not be reported as merged")
	}
}
```

- [ ] **Step 2: Run test and verify failure**

Run: `go test ./internal/git/ -run MergedInto`
Expected: FAIL — undefined.

- [ ] **Step 3: Add MergedInto to interface and implement**

Edit `internal/git/git.go`:

```go
type Git interface {
	IsRepo() bool
	ListBranches() ([]branch.Branch, error)
	AheadBehind(branch, base string) (ahead, behind int, err error)
	MergedInto(branch, base string) (bool, error)
}
```

Append to `internal/git/exec.go`:

```go
func (e *Exec) MergedInto(br, base string) (bool, error) {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", br, base)
	cmd.Dir = e.dir
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		return false, nil // documented "not an ancestor" exit code
	}
	return false, fmt.Errorf("git merge-base: %w", err)
}
```

- [ ] **Step 4: Verify git tests pass**

Run: `go test ./internal/git/ -v`
Expected: all PASS.

- [ ] **Step 5: Write failing test for the loader**

Create `internal/branch/load_test.go`:

```go
package branch

import (
	"errors"
	"testing"
	"time"
)

// fakeGit is a minimal Git stand-in for testing the loader.
type fakeGit struct {
	branches []Branch
	merged   map[string]bool // branch name → merged into base?
	ab       map[string][2]int
}

func (f *fakeGit) IsRepo() bool { return true }
func (f *fakeGit) ListBranches() ([]Branch, error) {
	cp := make([]Branch, len(f.branches))
	copy(cp, f.branches)
	return cp, nil
}
func (f *fakeGit) AheadBehind(b, base string) (int, int, error) {
	v, ok := f.ab[b]
	if !ok {
		return 0, 0, errors.New("no ab for " + b)
	}
	return v[0], v[1], nil
}
func (f *fakeGit) MergedInto(b, base string) (bool, error) {
	return f.merged[b], nil
}

func TestLoad_BasicFlow(t *testing.T) {
	t0 := time.Unix(1700000000, 0)
	f := &fakeGit{
		branches: []Branch{
			{Name: "main", IsCurrent: true, LastCommitTime: t0},
			{Name: "feature", Upstream: "origin/feature", Ahead: 1, LastCommitTime: t0},
			{Name: "stale", UpstreamGone: true, LastCommitTime: t0},
			{Name: "no-upstream", LastCommitTime: t0},
		},
		merged: map[string]bool{
			"feature":     false,
			"stale":       true,
			"no-upstream": false,
		},
		ab: map[string][2]int{
			"no-upstream": {2, 0}, // backfill candidate
		},
	}
	cfg := LoadConfig{
		BaseBranches:      []string{"main", "master"},
		ProtectedPatterns: []string{"release/*"},
	}
	bs, err := Load(f, cfg)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]Branch{}
	for _, b := range bs {
		byName[b.Name] = b
	}
	if !byName["main"].Protected {
		t.Error("main (base+current) should be protected")
	}
	if byName["stale"].Status() != StatusBoth {
		t.Errorf("stale should be StatusBoth, got %v", byName["stale"].Status())
	}
	if byName["no-upstream"].Ahead != 2 {
		t.Errorf("no-upstream Ahead should be backfilled to 2, got %d", byName["no-upstream"].Ahead)
	}
	if byName["feature"].MergedIntoBase {
		t.Error("feature should not be merged")
	}
}
```

- [ ] **Step 6: Run test and verify failure**

Run: `go test ./internal/branch/ -run Load`
Expected: FAIL — `Load` undefined.

- [ ] **Step 7: Implement loader**

Create `internal/branch/load.go`:

```go
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
```

- [ ] **Step 8: Run tests and verify pass**

Run: `go test ./internal/branch/ -v`
Expected: all PASS.

- [ ] **Step 9: Commit**

```bash
git add internal/git/git.go internal/git/exec.go internal/git/exec_test.go \
       internal/branch/load.go internal/branch/load_test.go
git commit -m "feat(branch): orchestrate branch load with classification and ab backfill"
```

---

## Task 10: Config loader

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

Default config when file missing. `XDG_CONFIG_HOME` respected. We auto-create with comments on first run.

- [ ] **Step 1: Add the TOML dependency**

```bash
go get github.com/BurntSushi/toml@latest
```

- [ ] **Step 2: Write failing tests**

Create `internal/config/config_test.go`:

```go
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
```

- [ ] **Step 3: Run tests and verify failure**

Run: `go test ./internal/config/`
Expected: FAIL — package empty.

- [ ] **Step 4: Implement**

Create `internal/config/config.go`:

```go
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
```

- [ ] **Step 5: Run tests and verify pass**

Run: `go test ./internal/config/ -v`
Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go go.mod go.sum
git commit -m "feat(config): load TOML config with XDG path and auto-create defaults"
```

---

## Task 11: `nathm list` subcommand

**Files:**
- Create: `cmd/list.go`
- Create: `cmd/list_test.go`

This is the first end-to-end exercise of the data path: real git, real config, real loader. No TUI. Useful for scripting too.

Output format: TSV. One branch per line. Columns: `name\tstatus\tage_seconds\tahead\tbehind`. Exit 1 if not in a git repo.

- [ ] **Step 1: Write failing test**

Create `cmd/list_test.go`:

```go
package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestList_Smoke(t *testing.T) {
	// Build the binary into the tempdir.
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "nathm")
	build := exec.Command("go", "build", "-o", bin, "github.com/USER/nathm")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}

	// Tempdir repo with two branches.
	repo := t.TempDir()
	for _, c := range [][]string{
		{"git", "init", "-q", "-b", "main"},
		{"git", "config", "user.email", "x@x"},
		{"git", "config", "user.name", "x"},
		{"git", "config", "commit.gpgsign", "false"},
		{"git", "commit", "--allow-empty", "-q", "-m", "init"},
		{"git", "branch", "feature"},
	} {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", c, err, out)
		}
	}

	// Isolate config to avoid touching the real user's config.
	cfgDir := filepath.Join(tmp, "cfg")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(bin, "list")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+cfgDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("nathm list: %v\nstderr: %s", err, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "main\t") || !strings.Contains(out, "feature\t") {
		t.Fatalf("missing branches in output: %q", out)
	}
}
```

- [ ] **Step 2: Run test and verify failure**

Run: `go test ./cmd/ -run TestList_Smoke`
Expected: FAIL — `list` subcommand doesn't exist.

- [ ] **Step 3: Implement the list subcommand**

Create `cmd/list.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/USER/nathm/internal/branch"
	"github.com/USER/nathm/internal/config"
	"github.com/USER/nathm/internal/git"
)

var (
	listStaleOnly bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Print local branches as TSV (name, status, age_seconds, ahead, behind)",
	RunE: func(cmd *cobra.Command, args []string) error {
		g := git.NewExec("")
		if !g.IsRepo() {
			return fmt.Errorf("not a git repository")
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		bs, err := branch.Load(g, branch.LoadConfig{
			BaseBranches:      cfg.BaseBranches,
			ProtectedPatterns: cfg.ProtectedPatterns,
		})
		if err != nil {
			return err
		}
		now := time.Now()
		w := cmd.OutOrStdout()
		for _, b := range bs {
			if listStaleOnly && !b.IsStale() {
				continue
			}
			age := int64(now.Sub(b.LastCommitTime).Seconds())
			fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\n", b.Name, b.Status(), age, b.Ahead, b.Behind)
		}
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVar(&listStaleOnly, "stale", false, "Only print stale (gone or merged) branches")
	rootCmd.AddCommand(listCmd)
}

// SilenceUsage on this command's parent already handles error printing in
// Execute(); list itself doesn't need to call os.Exit.
var _ = os.Stderr // keep os import for future use
```

If `os` is unused, drop the `_ = os.Stderr` line and the import.

- [ ] **Step 4: Run tests and verify pass**

Run: `go test ./cmd/ -v`
Expected: PASS. Note: this test builds the binary, so it's slower than other tests.

- [ ] **Step 5: Manual smoke test**

```bash
cd /Users/mohammadalamarneh/workspace/nathm
go run . list
```

Expected: TSV-formatted output of the current repo's branches (or "not a git repository" error).

- [ ] **Step 6: Commit**

```bash
git add cmd/list.go cmd/list_test.go
git commit -m "feat(cmd): add nathm list subcommand for scriptable branch listing"
```

---

## Task 12: Delete branch operation

**Files:**
- Modify: `internal/git/git.go`
- Modify: `internal/git/exec.go`
- Modify: `internal/git/exec_test.go`

Wraps `git branch -d` (safe) and `git branch -D` (force).

- [ ] **Step 1: Write failing test**

Append to `internal/git/exec_test.go`:

```go
func TestExec_DeleteBranch_Safe(t *testing.T) {
	dir := newTestRepo(t)
	runIn(t, dir, "git", "branch", "feature")
	g := NewExec(dir)
	if err := g.DeleteBranch("feature", false); err != nil {
		t.Fatalf("safe delete: %v", err)
	}
	bs, _ := g.ListBranches()
	for _, b := range bs {
		if b.Name == "feature" {
			t.Fatal("feature still exists after delete")
		}
	}
}

func TestExec_DeleteBranch_RefusesUnmerged(t *testing.T) {
	dir := newTestRepo(t)
	runIn(t, dir, "git", "checkout", "-q", "-b", "wip")
	writeFile(t, dir, "x.txt", "x")
	runIn(t, dir, "git", "add", "x.txt")
	runIn(t, dir, "git", "commit", "-q", "-m", "x")
	runIn(t, dir, "git", "checkout", "-q", "main")

	g := NewExec(dir)
	if err := g.DeleteBranch("wip", false); err == nil {
		t.Fatal("safe delete should refuse unmerged branch")
	}
}

func TestExec_DeleteBranch_Force(t *testing.T) {
	dir := newTestRepo(t)
	runIn(t, dir, "git", "checkout", "-q", "-b", "wip")
	writeFile(t, dir, "x.txt", "x")
	runIn(t, dir, "git", "add", "x.txt")
	runIn(t, dir, "git", "commit", "-q", "-m", "x")
	runIn(t, dir, "git", "checkout", "-q", "main")

	g := NewExec(dir)
	if err := g.DeleteBranch("wip", true); err != nil {
		t.Fatalf("force delete: %v", err)
	}
}
```

- [ ] **Step 2: Run test and verify failure**

Run: `go test ./internal/git/ -run DeleteBranch`
Expected: FAIL — undefined.

- [ ] **Step 3: Add to interface and implement**

Edit `internal/git/git.go`:

```go
type Git interface {
	IsRepo() bool
	ListBranches() ([]branch.Branch, error)
	AheadBehind(branch, base string) (ahead, behind int, err error)
	MergedInto(branch, base string) (bool, error)
	DeleteBranch(name string, force bool) error
}
```

Append to `internal/git/exec.go`:

```go
func (e *Exec) DeleteBranch(name string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	cmd := exec.Command("git", "branch", flag, name)
	cmd.Dir = e.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git branch %s %s: %w: %s", flag, name, err, strings.TrimSpace(string(out)))
	}
	return nil
}
```

- [ ] **Step 4: Run tests and verify pass**

Run: `go test ./internal/git/ -v`
Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/git/git.go internal/git/exec.go internal/git/exec_test.go
git commit -m "feat(git): delete branch with safe and force modes"
```

---

## Task 13: Rename branch operation

**Files:**
- Modify: `internal/git/git.go`
- Modify: `internal/git/exec.go`
- Modify: `internal/git/exec_test.go`

Wraps `git branch -m`. Refuses if target exists (git's default — we surface its error).

- [ ] **Step 1: Write failing test**

Append to `internal/git/exec_test.go`:

```go
func TestExec_RenameBranch(t *testing.T) {
	dir := newTestRepo(t)
	runIn(t, dir, "git", "branch", "old")
	g := NewExec(dir)
	if err := g.RenameBranch("old", "new"); err != nil {
		t.Fatalf("rename: %v", err)
	}
	bs, _ := g.ListBranches()
	names := map[string]bool{}
	for _, b := range bs {
		names[b.Name] = true
	}
	if !names["new"] || names["old"] {
		t.Fatalf("expected rename old→new, got %v", names)
	}
}

func TestExec_RenameBranch_TargetExists(t *testing.T) {
	dir := newTestRepo(t)
	runIn(t, dir, "git", "branch", "old")
	runIn(t, dir, "git", "branch", "new")
	g := NewExec(dir)
	if err := g.RenameBranch("old", "new"); err == nil {
		t.Fatal("expected error renaming to existing branch")
	}
}
```

- [ ] **Step 2: Run test and verify failure**

Run: `go test ./internal/git/ -run RenameBranch`
Expected: FAIL — undefined.

- [ ] **Step 3: Add to interface and implement**

Edit `internal/git/git.go`:

```go
type Git interface {
	IsRepo() bool
	ListBranches() ([]branch.Branch, error)
	AheadBehind(branch, base string) (ahead, behind int, err error)
	MergedInto(branch, base string) (bool, error)
	DeleteBranch(name string, force bool) error
	RenameBranch(oldName, newName string) error
}
```

Append to `internal/git/exec.go`:

```go
func (e *Exec) RenameBranch(oldName, newName string) error {
	cmd := exec.Command("git", "branch", "-m", oldName, newName)
	cmd.Dir = e.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git branch -m: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
```

- [ ] **Step 4: Run tests and verify pass**

Run: `go test ./internal/git/ -v`
Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/git/git.go internal/git/exec.go internal/git/exec_test.go
git commit -m "feat(git): rename local branch"
```

---

## Task 14: Checkout branch

**Files:**
- Modify: `internal/git/git.go`
- Modify: `internal/git/exec.go`
- Modify: `internal/git/exec_test.go`

- [ ] **Step 1: Write failing test**

Append to `internal/git/exec_test.go`:

```go
func TestExec_Checkout(t *testing.T) {
	dir := newTestRepo(t)
	runIn(t, dir, "git", "branch", "feature")
	g := NewExec(dir)
	if err := g.Checkout("feature"); err != nil {
		t.Fatalf("checkout: %v", err)
	}
	bs, _ := g.ListBranches()
	for _, b := range bs {
		if b.Name == "feature" && !b.IsCurrent {
			t.Fatal("feature should be current after checkout")
		}
	}
}
```

- [ ] **Step 2: Run test and verify failure**

Run: `go test ./internal/git/ -run TestExec_Checkout`
Expected: FAIL — undefined.

- [ ] **Step 3: Add to interface and implement**

Edit `internal/git/git.go`:

```go
type Git interface {
	IsRepo() bool
	ListBranches() ([]branch.Branch, error)
	AheadBehind(branch, base string) (ahead, behind int, err error)
	MergedInto(branch, base string) (bool, error)
	DeleteBranch(name string, force bool) error
	RenameBranch(oldName, newName string) error
	Checkout(name string) error
}
```

Append to `internal/git/exec.go`:

```go
func (e *Exec) Checkout(name string) error {
	cmd := exec.Command("git", "checkout", "-q", name)
	cmd.Dir = e.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
```

- [ ] **Step 4: Run tests and verify pass**

Run: `go test ./internal/git/ -v`
Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/git/git.go internal/git/exec.go internal/git/exec_test.go
git commit -m "feat(git): checkout branch"
```

---

## Task 15: Fetch with prune

**Files:**
- Modify: `internal/git/git.go`
- Modify: `internal/git/exec.go`
- Modify: `internal/git/exec_test.go`

`FetchPrune` runs `git fetch --all --prune`. Tested by configuring a local "remote" and observing `[gone]` in track output.

- [ ] **Step 1: Add testRepoWithRemote helper**

Append to `internal/git/testhelp_test.go`:

```go
// newTestRepoWithRemote creates a test repo with a real (local-path) remote
// and a single tracking branch "feature" whose upstream is later deleted on
// the remote — useful for testing "gone" detection.
func newTestRepoWithRemote(t *testing.T) (repo, remote string) {
	t.Helper()
	remote = t.TempDir()
	runIn(t, remote, "git", "init", "-q", "--bare", "-b", "main")
	repo = newTestRepo(t)
	runIn(t, repo, "git", "remote", "add", "origin", remote)
	runIn(t, repo, "git", "push", "-q", "-u", "origin", "main")
	runIn(t, repo, "git", "checkout", "-q", "-b", "feature")
	writeFile(t, repo, "f.txt", "f")
	runIn(t, repo, "git", "add", "f.txt")
	runIn(t, repo, "git", "commit", "-q", "-m", "f")
	runIn(t, repo, "git", "push", "-q", "-u", "origin", "feature")
	runIn(t, repo, "git", "checkout", "-q", "main")
	// Delete the remote branch so the next fetch --prune marks local "feature" as gone.
	runIn(t, remote, "git", "branch", "-D", "feature")
	return repo, remote
}
```

- [ ] **Step 2: Write failing test**

Append to `internal/git/exec_test.go`:

```go
func TestExec_FetchPrune_MarksGone(t *testing.T) {
	repo, _ := newTestRepoWithRemote(t)
	g := NewExec(repo)
	if err := g.FetchPrune(); err != nil {
		t.Fatalf("fetch: %v", err)
	}
	bs, err := g.ListBranches()
	if err != nil {
		t.Fatal(err)
	}
	var feature *Branch
	// (Branch lives in internal/branch; alias for clarity)
	type B = struct {
		Name         string
		UpstreamGone bool
	}
	_ = feature
	for i := range bs {
		if bs[i].Name == "feature" {
			if !bs[i].UpstreamGone {
				t.Fatalf("feature should be marked gone after fetch --prune; got %+v", bs[i])
			}
			return
		}
	}
	t.Fatal("feature branch missing")
}
```

- [ ] **Step 3: Run test and verify failure**

Run: `go test ./internal/git/ -run FetchPrune`
Expected: FAIL — undefined.

- [ ] **Step 4: Add to interface and implement**

Edit `internal/git/git.go`:

```go
type Git interface {
	IsRepo() bool
	ListBranches() ([]branch.Branch, error)
	AheadBehind(branch, base string) (ahead, behind int, err error)
	MergedInto(branch, base string) (bool, error)
	DeleteBranch(name string, force bool) error
	RenameBranch(oldName, newName string) error
	Checkout(name string) error
	FetchPrune() error
}
```

Append to `internal/git/exec.go`:

```go
func (e *Exec) FetchPrune() error {
	cmd := exec.Command("git", "fetch", "--all", "--prune", "-q")
	cmd.Dir = e.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch --prune: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
```

- [ ] **Step 5: Run tests and verify pass**

Run: `go test ./internal/git/ -v`
Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/git/git.go internal/git/exec.go internal/git/exec_test.go internal/git/testhelp_test.go
git commit -m "feat(git): fetch --prune to update gone tracking"
```

---

## Task 16: `nathm prune` subcommand

**Files:**
- Create: `cmd/prune.go`
- Create: `cmd/prune_test.go`

Flow: `fetch --prune` → reload branches → filter to stale + non-protected → print summary → ask for `y/N` (or skip if `--yes`) → delete each. Per-branch failures go to stderr; succeed if at least one was deleted; user cancellation exits 2.

- [ ] **Step 1: Write failing test**

Create `cmd/prune_test.go`:

```go
package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Runs `nathm prune --yes` against a tempdir repo with a known-stale
// (merged) branch, and verifies the branch was deleted.
func TestPrune_Yes_DeletesMergedBranch(t *testing.T) {
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "nathm")
	if out, err := exec.Command("go", "build", "-o", bin, "github.com/USER/nathm").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}

	repo := t.TempDir()
	for _, c := range [][]string{
		{"git", "init", "-q", "-b", "main"},
		{"git", "config", "user.email", "x@x"},
		{"git", "config", "user.name", "x"},
		{"git", "config", "commit.gpgsign", "false"},
		{"git", "commit", "--allow-empty", "-q", "-m", "init"},
		{"git", "checkout", "-q", "-b", "merged-feature"},
		{"git", "commit", "--allow-empty", "-q", "-m", "f"},
		{"git", "checkout", "-q", "main"},
		{"git", "merge", "--no-ff", "-q", "-m", "merge", "merged-feature"},
	} {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", c, err, out)
		}
	}

	cfgDir := filepath.Join(tmp, "cfg")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(bin, "prune", "--yes")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+cfgDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("prune: %v\nstderr: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "merged-feature") {
		t.Fatalf("expected merged-feature in output, got %q", stdout.String())
	}

	// Verify branch is gone.
	listCmd := exec.Command("git", "branch", "--list")
	listCmd.Dir = repo
	out, err := listCmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "merged-feature") {
		t.Fatalf("merged-feature still present:\n%s", out)
	}
}
```

- [ ] **Step 2: Run test and verify failure**

Run: `go test ./cmd/ -run TestPrune_Yes`
Expected: FAIL — `prune` subcommand doesn't exist.

- [ ] **Step 3: Implement**

Create `cmd/prune.go`:

```go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/USER/nathm/internal/branch"
	"github.com/USER/nathm/internal/config"
	"github.com/USER/nathm/internal/git"
)

var pruneYes bool

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Delete branches whose upstream is gone or which are merged into base",
	RunE: func(cmd *cobra.Command, args []string) error {
		g := git.NewExec("")
		if !g.IsRepo() {
			return fmt.Errorf("not a git repository")
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		// Refresh tracking state. Tolerate failures (offline, no remote).
		if err := g.FetchPrune(); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", err)
		}
		bs, err := branch.Load(g, branch.LoadConfig{
			BaseBranches:      cfg.BaseBranches,
			ProtectedPatterns: cfg.ProtectedPatterns,
		})
		if err != nil {
			return err
		}
		var candidates []branch.Branch
		for _, b := range bs {
			if b.IsStale() && !b.Protected {
				candidates = append(candidates, b)
			}
		}
		if len(candidates) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No stale branches to prune.")
			return nil
		}

		printCandidates(cmd.OutOrStdout(), candidates)

		if !pruneYes {
			confirmed, err := confirm(cmd.InOrStdin(), cmd.OutOrStdout())
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
				os.Exit(2)
			}
		}

		var deleted, failed int
		for _, b := range candidates {
			if err := g.DeleteBranch(b.Name, false); err != nil {
				// Try force delete only if branch was [gone] (work may live only on the now-deleted remote).
				if b.UpstreamGone {
					if err2 := g.DeleteBranch(b.Name, true); err2 == nil {
						deleted++
						fmt.Fprintf(cmd.OutOrStdout(), "deleted %s (force)\n", b.Name)
						continue
					}
				}
				failed++
				fmt.Fprintf(cmd.ErrOrStderr(), "failed: %s: %v\n", b.Name, err)
				continue
			}
			deleted++
			fmt.Fprintf(cmd.OutOrStdout(), "deleted %s\n", b.Name)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\nDone: %d deleted, %d failed.\n", deleted, failed)
		if deleted == 0 && failed > 0 {
			return fmt.Errorf("all deletions failed")
		}
		return nil
	},
}

func printCandidates(w *cobra.Command, candidates []branch.Branch) {
	// (signature actually takes io.Writer; helper inlined below)
}

func init() {
	pruneCmd.Flags().BoolVar(&pruneYes, "yes", false, "Skip confirmation prompt")
	rootCmd.AddCommand(pruneCmd)
}
```

The helper above had the wrong signature — replace `printCandidates` with the io.Writer version. Final cmd/prune.go:

```go
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/USER/nathm/internal/branch"
	"github.com/USER/nathm/internal/config"
	"github.com/USER/nathm/internal/git"
)

var pruneYes bool

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Delete branches whose upstream is gone or which are merged into base",
	RunE:  runPrune,
}

func runPrune(cmd *cobra.Command, args []string) error {
	g := git.NewExec("")
	if !g.IsRepo() {
		return fmt.Errorf("not a git repository")
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := g.FetchPrune(); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", err)
	}
	bs, err := branch.Load(g, branch.LoadConfig{
		BaseBranches:      cfg.BaseBranches,
		ProtectedPatterns: cfg.ProtectedPatterns,
	})
	if err != nil {
		return err
	}
	var candidates []branch.Branch
	for _, b := range bs {
		if b.IsStale() && !b.Protected {
			candidates = append(candidates, b)
		}
	}
	if len(candidates) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No stale branches to prune.")
		return nil
	}
	printCandidates(cmd.OutOrStdout(), candidates)

	if !pruneYes {
		ok, err := confirm(cmd.InOrStdin(), cmd.OutOrStdout())
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
			os.Exit(2)
		}
	}

	var deleted, failed int
	for _, b := range candidates {
		if err := g.DeleteBranch(b.Name, false); err != nil {
			if b.UpstreamGone {
				if err2 := g.DeleteBranch(b.Name, true); err2 == nil {
					deleted++
					fmt.Fprintf(cmd.OutOrStdout(), "deleted %s (force)\n", b.Name)
					continue
				}
			}
			failed++
			fmt.Fprintf(cmd.ErrOrStderr(), "failed: %s: %v\n", b.Name, err)
			continue
		}
		deleted++
		fmt.Fprintf(cmd.OutOrStdout(), "deleted %s\n", b.Name)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "\nDone: %d deleted, %d failed.\n", deleted, failed)
	if deleted == 0 && failed > 0 {
		return fmt.Errorf("all deletions failed")
	}
	return nil
}

func printCandidates(w io.Writer, candidates []branch.Branch) {
	fmt.Fprintf(w, "The following %d branch(es) will be deleted:\n", len(candidates))
	now := time.Now()
	for _, b := range candidates {
		age := humanize.RelTime(b.LastCommitTime, now, "ago", "from now")
		fmt.Fprintf(w, "  %-30s %-12s last commit %s\n", b.Name, b.Status(), age)
	}
}

func confirm(in io.Reader, out io.Writer) (bool, error) {
	fmt.Fprint(out, "Continue? [y/N] ")
	r := bufio.NewReader(in)
	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes", nil
}

func init() {
	pruneCmd.Flags().BoolVar(&pruneYes, "yes", false, "Skip confirmation prompt")
	rootCmd.AddCommand(pruneCmd)
}
```

- [ ] **Step 4: Add the humanize dependency**

```bash
go get github.com/dustin/go-humanize@latest
```

- [ ] **Step 5: Run tests and verify pass**

Run: `go test ./cmd/ -v`
Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add cmd/prune.go cmd/prune_test.go go.mod go.sum
git commit -m "feat(cmd): add nathm prune subcommand with --yes flag"
```

---

## Task 17: `nathm rename` subcommand

**Files:**
- Create: `cmd/rename.go`
- Create: `cmd/rename_test.go`

- [ ] **Step 1: Write failing test**

Create `cmd/rename_test.go`:

```go
package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRename_LocalBranch(t *testing.T) {
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "nathm")
	if out, err := exec.Command("go", "build", "-o", bin, "github.com/USER/nathm").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	repo := t.TempDir()
	for _, c := range [][]string{
		{"git", "init", "-q", "-b", "main"},
		{"git", "config", "user.email", "x@x"},
		{"git", "config", "user.name", "x"},
		{"git", "config", "commit.gpgsign", "false"},
		{"git", "commit", "--allow-empty", "-q", "-m", "init"},
		{"git", "branch", "old"},
	} {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", c, err, out)
		}
	}
	cfgDir := filepath.Join(tmp, "cfg")
	_ = os.MkdirAll(cfgDir, 0o755)

	cmd := exec.Command(bin, "rename", "old", "new")
	cmd.Dir = repo
	cmd.Env = append(os.Environ(), "XDG_CONFIG_HOME="+cfgDir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("rename: %v\n%s", err, stderr.String())
	}

	out, _ := exec.Command("git", "-C", repo, "branch", "--list").Output()
	if !strings.Contains(string(out), "new") || strings.Contains(string(out), "old") {
		t.Fatalf("expected old→new, got %s", out)
	}
}
```

- [ ] **Step 2: Run and verify failure**

Run: `go test ./cmd/ -run TestRename`
Expected: FAIL — subcommand missing.

- [ ] **Step 3: Implement**

Create `cmd/rename.go`:

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/USER/nathm/internal/git"
)

var renameCmd = &cobra.Command{
	Use:   "rename <old> <new>",
	Short: "Rename a local branch",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		g := git.NewExec("")
		if !g.IsRepo() {
			return fmt.Errorf("not a git repository")
		}
		return g.RenameBranch(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)
}
```

- [ ] **Step 4: Run tests and verify pass**

Run: `go test ./cmd/ -v`
Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/rename.go cmd/rename_test.go
git commit -m "feat(cmd): add nathm rename subcommand for local rename"
```

---

## Task 18: TUI scaffolding — model, view, list rendering (read-only)

**Files:**
- Create: `internal/tui/model.go`
- Create: `internal/tui/styles.go`
- Create: `internal/tui/keymap.go`
- Create: `internal/tui/messages.go`
- Create: `internal/tui/model_test.go`

First TUI task. Render the branch list using `bubbles/table` (no actions, no filter, no selection — those come next). Take branches as input rather than loading them — keeps the model testable.

- [ ] **Step 1: Add Bubble Tea dependencies**

```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/x/exp/teatest@latest
```

- [ ] **Step 2: Write failing test**

Create `internal/tui/model_test.go`:

```go
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
```

- [ ] **Step 3: Run and verify failure**

Run: `go test ./internal/tui/`
Expected: FAIL — `NewModel`, `SetSize` undefined.

- [ ] **Step 4: Implement minimal Model**

Create `internal/tui/styles.go`:

```go
package tui

import "github.com/charmbracelet/lipgloss"

var (
	statusActive    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	statusGone      = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	statusMerged    = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	statusBoth      = lipgloss.NewStyle().Foreground(lipgloss.Color("213"))
	statusProtected = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	currentMarker   = lipgloss.NewStyle().Bold(true)
	dim             = lipgloss.NewStyle().Faint(true)
)
```

Create `internal/tui/keymap.go`:

```go
package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Quit key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(key.WithKeys("q", "ctrl+c", "esc"), key.WithHelp("q", "quit")),
}
```

Create `internal/tui/messages.go`:

```go
package tui

// (will grow as we add async refresh in later tasks)
```

Create `internal/tui/model.go`:

```go
// Package tui implements the interactive branch-management view.
package tui

import (
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"

	"github.com/USER/nathm/internal/branch"
	"github.com/USER/nathm/internal/git"
)

// Model is the TUI's root.
type Model struct {
	branches []branch.Branch
	table    table.Model
	git      git.Git
	width    int
	height   int
	now      time.Time
	err      string
}

func NewModel(branches []branch.Branch, g git.Git) *Model {
	m := &Model{
		branches: branches,
		git:      g,
		now:      time.Now(),
	}
	m.rebuildTable()
	return m
}

func (m *Model) SetSize(w, h int) {
	m.width, m.height = w, h
	m.table.SetWidth(w)
	m.table.SetHeight(maxInt(h-3, 5))
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	header := lipgloss.NewStyle().Bold(true).Render("nathm — local branches")
	footer := dim.Render("q quit · (more keys coming)")
	if m.err != "" {
		footer = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(m.err)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, m.table.View(), footer)
}

func (m *Model) rebuildTable() {
	cols := []table.Column{
		{Title: "", Width: 2},   // current marker
		{Title: "Branch", Width: 32},
		{Title: "Status", Width: 12},
		{Title: "Age", Width: 12},
		{Title: "↑ ↓", Width: 8},
		{Title: "Last commit", Width: 40},
	}
	rows := make([]table.Row, 0, len(m.branches))
	for _, b := range branchesSorted(m.branches) {
		marker := " "
		if b.IsCurrent {
			marker = "*"
		}
		rows = append(rows, table.Row{
			marker,
			b.Name,
			statusLabel(b),
			humanize.RelTime(b.LastCommitTime, m.now, "", "from now"),
			fmt.Sprintf("↑%d ↓%d", b.Ahead, b.Behind),
			truncate(b.LastCommitSubject, 40),
		})
	}
	m.table = table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
	)
	if m.width > 0 {
		m.SetSize(m.width, m.height)
	}
}

// branchesSorted: stale-first, then by last-commit age desc.
func branchesSorted(in []branch.Branch) []branch.Branch {
	out := make([]branch.Branch, len(in))
	copy(out, in)
	sort.SliceStable(out, func(i, j int) bool {
		si, sj := stalePriority(out[i]), stalePriority(out[j])
		if si != sj {
			return si > sj
		}
		// older = older time = smaller — but we want oldest first, so:
		return out[i].LastCommitTime.Before(out[j].LastCommitTime)
	})
	return out
}

func stalePriority(b branch.Branch) int {
	switch b.Status() {
	case branch.StatusBoth:
		return 3
	case branch.StatusGone:
		return 2
	case branch.StatusMerged:
		return 1
	default:
		return 0
	}
}

func statusLabel(b branch.Branch) string {
	if b.Protected {
		return "protected"
	}
	return b.Status().String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
```

- [ ] **Step 5: Run tests and verify pass**

Run: `go test ./internal/tui/ -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/tui/ go.mod go.sum
git commit -m "feat(tui): scaffold model with read-only branch list rendering"
```

---

## Task 19: TUI navigation and multi-select

**Files:**
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/keymap.go`
- Modify: `internal/tui/model_test.go`

`bubbles/table` already handles up/down navigation. We add multi-select tracking on top.

- [ ] **Step 1: Write failing test**

Append to `internal/tui/model_test.go`:

```go
import (
	// add:
	tea "github.com/charmbracelet/bubbletea"
)

func TestModel_Selection_ToggleAndClearAll(t *testing.T) {
	bs := []branch.Branch{
		{Name: "a", LastCommitTime: time.Now()},
		{Name: "b", LastCommitTime: time.Now()},
		{Name: "c", LastCommitTime: time.Now()},
	}
	m := NewModel(bs, nil)
	m.SetSize(120, 30)

	// space toggles current row
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !m.IsSelected(m.CurrentName()) {
		t.Fatal("space should select current branch")
	}
	// space again deselects
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if m.IsSelected(m.CurrentName()) {
		t.Fatal("space should deselect")
	}
	// 'a' selects all visible
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if len(m.Selected()) != 3 {
		t.Fatalf("expected 3 selected, got %d", len(m.Selected()))
	}
	// 'A' clears
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	if len(m.Selected()) != 0 {
		t.Fatalf("expected 0 selected after clear, got %d", len(m.Selected()))
	}
}

func updateModel(m *Model, msg tea.Msg) (*Model, tea.Cmd) {
	mm, cmd := m.Update(msg)
	return mm.(*Model), cmd
}
```

- [ ] **Step 2: Run and verify failure**

Run: `go test ./internal/tui/ -run Selection`
Expected: FAIL — `IsSelected`, `CurrentName`, `Selected` undefined.

- [ ] **Step 3: Implement**

Edit `internal/tui/keymap.go` — add bindings:

```go
package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Quit       key.Binding
	Select     key.Binding
	SelectAll  key.Binding
	ClearAll   key.Binding
}

var keys = keyMap{
	Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c", "esc"), key.WithHelp("q", "quit")),
	Select:    key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "select")),
	SelectAll: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "select all")),
	ClearAll:  key.NewBinding(key.WithKeys("A"), key.WithHelp("A", "clear selection")),
}
```

Modify `internal/tui/model.go`:

Add a `selected map[string]bool` field on Model:

```go
type Model struct {
	branches []branch.Branch
	table    table.Model
	git      git.Git
	width    int
	height   int
	now      time.Time
	err      string
	selected map[string]bool
}
```

Initialize it in `NewModel`:

```go
func NewModel(branches []branch.Branch, g git.Git) *Model {
	m := &Model{
		branches: branches,
		git:      g,
		now:      time.Now(),
		selected: make(map[string]bool),
	}
	m.rebuildTable()
	return m
}
```

Add the helper methods:

```go
func (m *Model) CurrentName() string {
	row := m.table.SelectedRow()
	if len(row) < 2 {
		return ""
	}
	return row[1]
}

func (m *Model) IsSelected(name string) bool {
	return m.selected[name]
}

func (m *Model) Selected() []string {
	out := make([]string, 0, len(m.selected))
	for name, ok := range m.selected {
		if ok {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

func (m *Model) toggleSelect(name string) {
	if name == "" {
		return
	}
	if m.selected[name] {
		delete(m.selected, name)
	} else {
		m.selected[name] = true
	}
}

func (m *Model) selectAllVisible() {
	for _, b := range m.branches {
		if !b.Protected && !b.IsCurrent {
			m.selected[b.Name] = true
		}
	}
}

func (m *Model) clearSelection() {
	m.selected = map[string]bool{}
}
```

Wire keys into `Update`. Replace the `KeyMsg` case:

```go
case tea.KeyMsg:
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, keys.Select):
		m.toggleSelect(m.CurrentName())
		m.rebuildTable()
		return m, nil
	case key.Matches(msg, keys.SelectAll):
		m.selectAllVisible()
		m.rebuildTable()
		return m, nil
	case key.Matches(msg, keys.ClearAll):
		m.clearSelection()
		m.rebuildTable()
		return m, nil
	}
```

Add `key` import: `"github.com/charmbracelet/bubbles/key"`.

Update `rebuildTable` to render selection markers — change the first column to show `[x]` for selected, `[ ]` otherwise, and combine with the current marker:

```go
cols := []table.Column{
    {Title: "Sel", Width: 4},
    {Title: "Branch", Width: 32},
    // ... rest unchanged
}
rows := make([]table.Row, 0, len(m.branches))
for _, b := range branchesSorted(m.branches) {
    marker := "[ ]"
    if m.selected[b.Name] {
        marker = "[x]"
    }
    if b.IsCurrent {
        marker = " * "
    }
    rows = append(rows, table.Row{
        marker,
        b.Name,
        statusLabel(b),
        humanize.RelTime(b.LastCommitTime, m.now, "", "from now"),
        fmt.Sprintf("↑%d ↓%d", b.Ahead, b.Behind),
        truncate(b.LastCommitSubject, 40),
    })
}
```

- [ ] **Step 4: Run tests and verify pass**

Run: `go test ./internal/tui/ -v`
Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/tui/
git commit -m "feat(tui): multi-select with space/a/A keybindings"
```

---

## Task 20: TUI filter and sort

**Files:**
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/keymap.go`
- Modify: `internal/tui/model_test.go`

`/` opens an inline filter input; typing filters by substring. `s` cycles sort order. `p` toggles "stale only".

- [ ] **Step 1: Add textinput dependency (already in bubbles)**

No new dep needed — `github.com/charmbracelet/bubbles/textinput` ships with the `bubbles` module.

- [ ] **Step 2: Write failing test**

Append to `internal/tui/model_test.go`:

```go
func TestModel_Filter_ByName(t *testing.T) {
	bs := []branch.Branch{
		{Name: "feature-a", LastCommitTime: time.Now()},
		{Name: "feature-b", LastCommitTime: time.Now()},
		{Name: "main", IsCurrent: true, LastCommitTime: time.Now()},
	}
	m := NewModel(bs, nil)
	m.SetSize(120, 30)
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	for _, r := range "feature" {
		m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})
	out := m.View()
	if !strings.Contains(out, "feature-a") || !strings.Contains(out, "feature-b") {
		t.Errorf("expected feature-* visible:\n%s", out)
	}
	if strings.Contains(out, "main") {
		t.Errorf("main should be filtered out:\n%s", out)
	}
}

func TestModel_StaleOnlyToggle(t *testing.T) {
	bs := []branch.Branch{
		{Name: "active", LastCommitTime: time.Now()},
		{Name: "gone-1", UpstreamGone: true, LastCommitTime: time.Now()},
	}
	m := NewModel(bs, nil)
	m.SetSize(120, 30)
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	out := m.View()
	if strings.Contains(out, "active") {
		t.Errorf("active should be hidden in stale-only mode:\n%s", out)
	}
	if !strings.Contains(out, "gone-1") {
		t.Errorf("gone-1 should still be shown:\n%s", out)
	}
}
```

- [ ] **Step 3: Run and verify failure**

Run: `go test ./internal/tui/ -run "Filter|StaleOnly"`
Expected: FAIL.

- [ ] **Step 4: Implement**

Edit `internal/tui/keymap.go` — add bindings:

```go
type keyMap struct {
	Quit       key.Binding
	Select     key.Binding
	SelectAll  key.Binding
	ClearAll   key.Binding
	Filter     key.Binding
	Sort       key.Binding
	StaleOnly  key.Binding
}

var keys = keyMap{
	Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Select:    key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "select")),
	SelectAll: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "select all")),
	ClearAll:  key.NewBinding(key.WithKeys("A"), key.WithHelp("A", "clear selection")),
	Filter:    key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
	Sort:      key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "cycle sort")),
	StaleOnly: key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "stale only")),
}
```

Note: removed `esc` from `Quit` because it now means "exit filter mode".

Modify `internal/tui/model.go`:

```go
import (
	// add:
	"github.com/charmbracelet/bubbles/textinput"
	"strings"
)

type sortMode int

const (
	sortStaleFirst sortMode = iota
	sortName
	sortAge
)

type Model struct {
	branches      []branch.Branch
	table         table.Model
	git           git.Git
	width, height int
	now           time.Time
	err           string
	selected      map[string]bool

	filter      textinput.Model
	filterOn    bool
	filterText  string
	sortMode    sortMode
	staleOnly   bool
}

func NewModel(branches []branch.Branch, g git.Git) *Model {
	ti := textinput.New()
	ti.Placeholder = "filter..."
	ti.CharLimit = 80
	ti.Width = 40
	m := &Model{
		branches: branches,
		git:      g,
		now:      time.Now(),
		selected: make(map[string]bool),
		filter:   ti,
	}
	m.rebuildTable()
	return m
}
```

Replace the `Update` body:

```go
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		if m.filterOn {
			switch msg.Type {
			case tea.KeyEnter, tea.KeyEsc:
				m.filterOn = false
				m.filter.Blur()
				if msg.Type == tea.KeyEsc {
					m.filterText = ""
					m.filter.SetValue("")
				} else {
					m.filterText = m.filter.Value()
				}
				m.rebuildTable()
				return m, nil
			}
			var cmd tea.Cmd
			m.filter, cmd = m.filter.Update(msg)
			return m, cmd
		}
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Select):
			m.toggleSelect(m.CurrentName())
			m.rebuildTable()
			return m, nil
		case key.Matches(msg, keys.SelectAll):
			m.selectAllVisible()
			m.rebuildTable()
			return m, nil
		case key.Matches(msg, keys.ClearAll):
			m.clearSelection()
			m.rebuildTable()
			return m, nil
		case key.Matches(msg, keys.Filter):
			m.filterOn = true
			m.filter.SetValue(m.filterText)
			m.filter.Focus()
			return m, nil
		case key.Matches(msg, keys.Sort):
			m.sortMode = (m.sortMode + 1) % 3
			m.rebuildTable()
			return m, nil
		case key.Matches(msg, keys.StaleOnly):
			m.staleOnly = !m.staleOnly
			m.rebuildTable()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}
```

Refactor visible-row computation. Add a `visibleBranches()` method:

```go
func (m *Model) visibleBranches() []branch.Branch {
	out := m.branches[:0:0] // new slice
	for _, b := range m.branches {
		if m.staleOnly && !b.IsStale() {
			continue
		}
		if m.filterText != "" && !strings.Contains(b.Name, m.filterText) {
			continue
		}
		out = append(out, b)
	}
	return m.applySort(out)
}

func (m *Model) applySort(in []branch.Branch) []branch.Branch {
	out := make([]branch.Branch, len(in))
	copy(out, in)
	switch m.sortMode {
	case sortName:
		sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	case sortAge:
		sort.SliceStable(out, func(i, j int) bool {
			return out[i].LastCommitTime.Before(out[j].LastCommitTime)
		})
	default: // sortStaleFirst
		sort.SliceStable(out, func(i, j int) bool {
			si, sj := stalePriority(out[i]), stalePriority(out[j])
			if si != sj {
				return si > sj
			}
			return out[i].LastCommitTime.Before(out[j].LastCommitTime)
		})
	}
	return out
}
```

Replace `branchesSorted` calls in `rebuildTable` with `m.visibleBranches()`:

```go
for _, b := range m.visibleBranches() {
    // ... existing row construction
}
```

(Delete the now-unused `branchesSorted` function.)

Update `View` to show the filter input when active:

```go
func (m *Model) View() string {
	header := lipgloss.NewStyle().Bold(true).Render("nathm — local branches")
	mid := m.table.View()
	footer := dim.Render("space:select / a:all / A:clear / /:filter / s:sort / p:stale-only / q:quit")
	if m.filterOn {
		footer = m.filter.View()
	}
	if m.err != "" {
		footer = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(m.err)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, mid, footer)
}
```

- [ ] **Step 5: Run tests and verify pass**

Run: `go test ./internal/tui/ -v`
Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/tui/
git commit -m "feat(tui): filter, cycle sort, and stale-only toggle"
```

---

## Task 21: TUI delete with confirm dialog

**Files:**
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/keymap.go`
- Modify: `internal/tui/model_test.go`

`d` (safe) and `D` (force) trigger a confirm modal that lists which branches will be acted on (cursor row, or all selected if any). Pressing `y` performs deletes; anything else cancels.

- [ ] **Step 1: Write failing test**

Append to `internal/tui/model_test.go`:

```go
// fakeGit lets the model run without a real repo.
type tuiFakeGit struct {
	deleted []string
}

func (f *tuiFakeGit) IsRepo() bool                                     { return true }
func (f *tuiFakeGit) ListBranches() ([]branch.Branch, error)           { return nil, nil }
func (f *tuiFakeGit) AheadBehind(string, string) (int, int, error)     { return 0, 0, nil }
func (f *tuiFakeGit) MergedInto(string, string) (bool, error)          { return false, nil }
func (f *tuiFakeGit) DeleteBranch(name string, force bool) error {
	f.deleted = append(f.deleted, name)
	return nil
}
func (f *tuiFakeGit) RenameBranch(string, string) error { return nil }
func (f *tuiFakeGit) Checkout(string) error             { return nil }
func (f *tuiFakeGit) FetchPrune() error                 { return nil }

func TestModel_DeleteFlow(t *testing.T) {
	bs := []branch.Branch{
		{Name: "feature", LastCommitTime: time.Now()},
		{Name: "main", IsCurrent: true, Protected: true, LastCommitTime: time.Now()},
	}
	g := &tuiFakeGit{}
	m := NewModel(bs, g)
	m.SetSize(120, 30)
	// Move cursor to "feature" if not already.
	// Press d.
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if !m.Confirming() {
		t.Fatal("expected confirm modal active")
	}
	// Press y — should call DeleteBranch.
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if len(g.deleted) != 1 || g.deleted[0] != "feature" {
		t.Fatalf("expected delete of feature, got %v", g.deleted)
	}
}
```

- [ ] **Step 2: Run and verify failure**

Run: `go test ./internal/tui/ -run DeleteFlow`
Expected: FAIL.

- [ ] **Step 3: Implement confirm modal + delete action**

Edit `internal/tui/keymap.go`:

```go
type keyMap struct {
	Quit         key.Binding
	Select       key.Binding
	SelectAll    key.Binding
	ClearAll     key.Binding
	Filter       key.Binding
	Sort         key.Binding
	StaleOnly    key.Binding
	Delete       key.Binding
	ForceDelete  key.Binding
}

var keys = keyMap{
	Quit:        key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Select:      key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "select")),
	SelectAll:   key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "select all")),
	ClearAll:    key.NewBinding(key.WithKeys("A"), key.WithHelp("A", "clear selection")),
	Filter:      key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
	Sort:        key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "cycle sort")),
	StaleOnly:   key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "stale only")),
	Delete:      key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	ForceDelete: key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "force delete")),
}
```

Modify `internal/tui/model.go`. Add a confirm-state field:

```go
type confirmKind int

const (
	confirmNone confirmKind = iota
	confirmDelete
	confirmForceDelete
)

type Model struct {
	// ... existing fields
	confirmKind    confirmKind
	confirmTargets []string
}

func (m *Model) Confirming() bool { return m.confirmKind != confirmNone }
```

In `Update`, intercept keys when confirming:

```go
case tea.KeyMsg:
	if m.Confirming() {
		switch msg.String() {
		case "y", "Y":
			m.runConfirmedAction()
			return m, nil
		default:
			m.confirmKind = confirmNone
			m.confirmTargets = nil
			return m, nil
		}
	}
	if m.filterOn { /* unchanged */ }
	switch {
	// ... existing cases
	case key.Matches(msg, keys.Delete):
		m.beginDelete(false)
		return m, nil
	case key.Matches(msg, keys.ForceDelete):
		m.beginDelete(true)
		return m, nil
	}
```

Add the helpers:

```go
func (m *Model) targetsForAction() []string {
	sel := m.Selected()
	if len(sel) > 0 {
		return sel
	}
	if name := m.CurrentName(); name != "" {
		return []string{name}
	}
	return nil
}

func (m *Model) beginDelete(force bool) {
	targets := m.targetsForAction()
	// drop protected
	final := targets[:0:0]
	for _, name := range targets {
		if b, ok := m.byName(name); ok && !b.Protected {
			final = append(final, name)
		}
	}
	if len(final) == 0 {
		m.err = "nothing to delete (all targets protected)"
		return
	}
	m.confirmTargets = final
	if force {
		m.confirmKind = confirmForceDelete
	} else {
		m.confirmKind = confirmDelete
	}
}

func (m *Model) byName(name string) (branch.Branch, bool) {
	for _, b := range m.branches {
		if b.Name == name {
			return b, true
		}
	}
	return branch.Branch{}, false
}

func (m *Model) runConfirmedAction() {
	force := m.confirmKind == confirmForceDelete
	var failed []string
	for _, name := range m.confirmTargets {
		if err := m.git.DeleteBranch(name, force); err != nil {
			failed = append(failed, name+": "+err.Error())
		}
	}
	// Drop deleted from the in-memory list so the UI updates.
	keep := m.branches[:0:0]
	deleted := map[string]bool{}
	for _, n := range m.confirmTargets {
		deleted[n] = true
	}
	for _, b := range m.branches {
		if !deleted[b.Name] {
			keep = append(keep, b)
		}
	}
	m.branches = keep
	m.clearSelection()
	if len(failed) > 0 {
		m.err = "errors: " + strings.Join(failed, "; ")
	} else {
		m.err = ""
	}
	m.confirmKind = confirmNone
	m.confirmTargets = nil
	m.rebuildTable()
}
```

Update `View` to render the confirm modal when active:

```go
func (m *Model) View() string {
	header := lipgloss.NewStyle().Bold(true).Render("nathm — local branches")
	mid := m.table.View()
	footer := dim.Render("space:select / a:all / A:clear / /:filter / s:sort / p:stale-only / d:del / D:force / q:quit")
	if m.filterOn {
		footer = m.filter.View()
	}
	if m.err != "" {
		footer = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(m.err)
	}
	if m.Confirming() {
		mid = m.renderConfirm()
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, mid, footer)
}

func (m *Model) renderConfirm() string {
	verb := "Delete"
	if m.confirmKind == confirmForceDelete {
		verb = "FORCE DELETE"
	}
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("203")).
		Render(fmt.Sprintf("%s %d branch(es)?", verb, len(m.confirmTargets)))
	body := strings.Join(m.confirmTargets, "\n  ")
	help := dim.Render("y to confirm · any other key to cancel")
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
	return box.Render(lipgloss.JoinVertical(lipgloss.Left, title, "  "+body, "", help))
}
```

- [ ] **Step 4: Run tests and verify pass**

Run: `go test ./internal/tui/ -v`
Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/tui/
git commit -m "feat(tui): delete and force-delete with confirm modal"
```

---

## Task 22: TUI rename

**Files:**
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/keymap.go`
- Modify: `internal/tui/model_test.go`

`r` opens a rename input prefilled with the cursor branch's name. Enter commits, esc cancels. Never bulk.

- [ ] **Step 1: Write failing test**

Append to `internal/tui/model_test.go`:

```go
func (f *tuiFakeGit) renames() [][2]string { return f.renamed }

// Add to tuiFakeGit:
//   renamed [][2]string
// And implement RenameBranch to record.
```

Modify the existing `tuiFakeGit` definition:

```go
type tuiFakeGit struct {
	deleted []string
	renamed [][2]string
}

func (f *tuiFakeGit) RenameBranch(oldN, newN string) error {
	f.renamed = append(f.renamed, [2]string{oldN, newN})
	return nil
}
```

Then the test:

```go
func TestModel_RenameFlow(t *testing.T) {
	bs := []branch.Branch{
		{Name: "old-name", LastCommitTime: time.Now()},
	}
	g := &tuiFakeGit{}
	m := NewModel(bs, g)
	m.SetSize(120, 30)

	// press r to start rename
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if !m.Renaming() {
		t.Fatal("expected rename mode")
	}
	// clear and type new name
	for _, r := range "new-name" {
		m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	// Replace the prefilled value: simpler — just call internal helper that's exposed, OR
	// use Backspace to clear (we'll trust pre-fill; better: blank input then type).

	// For determinism, set the value directly via the public-test setter:
	m.SetRenameValue("new-name")
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})

	if len(g.renamed) != 1 || g.renamed[0] != [2]string{"old-name", "new-name"} {
		t.Fatalf("expected rename old→new, got %v", g.renamed)
	}
}
```

- [ ] **Step 2: Run and verify failure**

Run: `go test ./internal/tui/ -run RenameFlow`
Expected: FAIL.

- [ ] **Step 3: Implement**

Edit `internal/tui/keymap.go` — add Rename binding:

```go
Rename: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rename")),
```

Edit `internal/tui/model.go`:

```go
type Model struct {
	// ... existing fields
	renameOn     bool
	renameSource string
	renameInput  textinput.Model
}
```

Initialize in `NewModel`:

```go
ri := textinput.New()
ri.CharLimit = 200
ri.Width = 40
m := &Model{
	// ... existing
	renameInput: ri,
}
```

Add helper:

```go
func (m *Model) Renaming() bool { return m.renameOn }

func (m *Model) SetRenameValue(s string) {
	m.renameInput.SetValue(s)
}

func (m *Model) beginRename() {
	name := m.CurrentName()
	if name == "" {
		return
	}
	if b, ok := m.byName(name); ok && b.Protected {
		m.err = "cannot rename protected branch"
		return
	}
	m.renameOn = true
	m.renameSource = name
	m.renameInput.SetValue(name)
	m.renameInput.Focus()
}

func (m *Model) commitRename() {
	new := strings.TrimSpace(m.renameInput.Value())
	if new == "" || new == m.renameSource {
		m.cancelRename()
		return
	}
	if err := m.git.RenameBranch(m.renameSource, new); err != nil {
		m.err = "rename failed: " + err.Error()
		m.cancelRename()
		return
	}
	for i := range m.branches {
		if m.branches[i].Name == m.renameSource {
			m.branches[i].Name = new
		}
	}
	m.cancelRename()
	m.rebuildTable()
}

func (m *Model) cancelRename() {
	m.renameOn = false
	m.renameSource = ""
	m.renameInput.SetValue("")
	m.renameInput.Blur()
}
```

In `Update`, before the confirm check, add:

```go
case tea.KeyMsg:
	if m.renameOn {
		switch msg.Type {
		case tea.KeyEsc:
			m.cancelRename()
			return m, nil
		case tea.KeyEnter:
			m.commitRename()
			return m, nil
		}
		var cmd tea.Cmd
		m.renameInput, cmd = m.renameInput.Update(msg)
		return m, cmd
	}
	// ... existing logic
```

Add the case in the keymap switch:

```go
case key.Matches(msg, keys.Rename):
	m.beginRename()
	return m, nil
```

Update `View` to show the rename input:

```go
if m.renameOn {
	footer = "rename: " + m.renameInput.View() + "  (enter:save · esc:cancel)"
}
```

- [ ] **Step 4: Run tests and verify pass**

Run: `go test ./internal/tui/ -v`
Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/tui/
git commit -m "feat(tui): inline rename input with enter to save"
```

---

## Task 23: TUI checkout

**Files:**
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/keymap.go`
- Modify: `internal/tui/model_test.go`

`c` checks out the cursor branch. After success, mark it current, refresh.

- [ ] **Step 1: Write failing test**

Append to `internal/tui/model_test.go`:

```go
func (f *tuiFakeGit) checkedOut() []string { return f.checkout }

// Add to tuiFakeGit:
//   checkout []string

func TestModel_CheckoutFlow(t *testing.T) {
	bs := []branch.Branch{
		{Name: "main", IsCurrent: true, Protected: true, LastCommitTime: time.Now()},
		{Name: "feature", LastCommitTime: time.Now()},
	}
	g := &tuiFakeGit{}
	m := NewModel(bs, g)
	m.SetSize(120, 30)
	// Move cursor to feature.
	m.SetCursorByName("feature")
	m, _ = updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

	if len(g.checkout) != 1 || g.checkout[0] != "feature" {
		t.Fatalf("expected checkout of feature, got %v", g.checkout)
	}
}
```

Modify `tuiFakeGit`:

```go
type tuiFakeGit struct {
	deleted  []string
	renamed  [][2]string
	checkout []string
}

func (f *tuiFakeGit) Checkout(name string) error {
	f.checkout = append(f.checkout, name)
	return nil
}
```

- [ ] **Step 2: Run and verify failure**

Run: `go test ./internal/tui/ -run CheckoutFlow`
Expected: FAIL.

- [ ] **Step 3: Implement**

Edit `internal/tui/keymap.go`:

```go
Checkout: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "checkout")),
```

Edit `internal/tui/model.go`. Add the `SetCursorByName` helper for tests:

```go
func (m *Model) SetCursorByName(name string) {
	visible := m.visibleBranches()
	for i, b := range visible {
		if b.Name == name {
			m.table.SetCursor(i)
			return
		}
	}
}
```

Add `doCheckout`:

```go
func (m *Model) doCheckout() {
	name := m.CurrentName()
	if name == "" {
		return
	}
	if err := m.git.Checkout(name); err != nil {
		m.err = "checkout failed: " + err.Error()
		return
	}
	for i := range m.branches {
		m.branches[i].IsCurrent = (m.branches[i].Name == name)
	}
	m.err = ""
	m.rebuildTable()
}
```

Wire the key:

```go
case key.Matches(msg, keys.Checkout):
	m.doCheckout()
	return m, nil
```

- [ ] **Step 4: Run tests and verify pass**

Run: `go test ./internal/tui/ -v`
Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/tui/
git commit -m "feat(tui): checkout branch with c key"
```

---

## Task 24: TUI status bar polish + help overlay

**Files:**
- Modify: `internal/tui/model.go`
- Modify: `internal/tui/keymap.go`

A `?` key toggles a full keymap overlay. The status bar shows a transient last-action message.

- [ ] **Step 1: Add Help binding**

Edit `internal/tui/keymap.go`:

```go
Help: key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
```

- [ ] **Step 2: Implement**

Edit `internal/tui/model.go`. Add field:

```go
showHelp bool
```

Toggle in Update:

```go
case key.Matches(msg, keys.Help):
	m.showHelp = !m.showHelp
	return m, nil
```

Add render method:

```go
func (m *Model) renderHelp() string {
	lines := []string{
		"nathm — keybindings",
		"",
		"  ↑/↓ or j/k    navigate",
		"  space         toggle selection",
		"  a             select all visible",
		"  A             clear selection",
		"  enter / d     delete (cursor or selected)",
		"  D             force delete",
		"  r             rename (cursor only)",
		"  c             checkout",
		"  /             filter by name",
		"  s             cycle sort",
		"  p             toggle stale-only",
		"  ?             toggle this help",
		"  q             quit",
	}
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
	return box.Render(strings.Join(lines, "\n"))
}
```

In `View`:

```go
if m.showHelp {
	mid = m.renderHelp()
}
```

(Help wins over the table view but not over Confirming/Renaming — order matters; place after the existing conditionals if you want help to be the highest-priority overlay, or before to let confirms win. Recommended: confirm wins over help.)

- [ ] **Step 3: Smoke test manually**

Run: `go run . list` to check nothing broke. (TUI itself is wired in Task 25.)

- [ ] **Step 4: Commit**

```bash
git add internal/tui/
git commit -m "feat(tui): help overlay (?)"
```

---

## Task 25: Wire root command to launch the TUI

**Files:**
- Modify: `cmd/root.go`

- [ ] **Step 1: Replace the placeholder RunE**

Edit `cmd/root.go`:

```go
package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/USER/nathm/internal/branch"
	"github.com/USER/nathm/internal/config"
	"github.com/USER/nathm/internal/git"
	"github.com/USER/nathm/internal/tui"
)

var rootCmd = &cobra.Command{
	Use:   "nathm",
	Short: "Organize and clean up local git branches",
	Long:  "nathm (نظم) is an interactive TUI and CLI for organizing local git branches.",
	RunE:  runRoot,
}

func runRoot(cmd *cobra.Command, args []string) error {
	g := git.NewExec("")
	if !g.IsRepo() {
		return fmt.Errorf("not a git repository")
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	bs, err := branch.Load(g, branch.LoadConfig{
		BaseBranches:      cfg.BaseBranches,
		ProtectedPatterns: cfg.ProtectedPatterns,
	})
	if err != nil {
		return err
	}
	m := tui.NewModel(bs, g)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./... && ls -l nathm`
Expected: binary exists; no compile errors.

- [ ] **Step 3: Manual smoke test**

In a real git repo (or one with multiple branches):

```bash
cd /some/git/repo
/Users/mohammadalamarneh/workspace/nathm/nathm
```

Verify: TUI opens, branches list, j/k navigates, q quits cleanly. Try `/`, `s`, `p`. Don't try delete on anything important.

- [ ] **Step 4: Commit**

```bash
git add cmd/root.go
git commit -m "feat(cmd): launch TUI from default invocation"
```

---

## Task 26: README + final manual verification

**Files:**
- Create: `README.md`

- [ ] **Step 1: Write README**

Create `README.md`:

```markdown
# nathm

`nathm` (Arabic نظم — *to organize*) is a TUI-first CLI for organizing and cleaning up local git branches.

## Install

```bash
go install github.com/USER/nathm@latest
```

(Replace `USER` with the GitHub user/org once published.)

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
| `d` | safe delete (cursor or selected) |
| `D` | force delete |
| `r` | rename (cursor only) |
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

## Status

v0.1: TUI + prune + rename + list. GitHub PR awareness and remote rename are deferred to v2.
```

- [ ] **Step 2: Run the full test suite one more time**

Run: `go test ./...`
Expected: all PASS.

- [ ] **Step 3: Build a release binary**

Run: `go build -o nathm .`
Expected: produces `nathm` in repo root.

- [ ] **Step 4: Manual end-to-end smoke**

Pick a real-but-low-stakes git repo (e.g. clone something throwaway) and exercise:

1. `./nathm version` → prints version.
2. `./nathm list` → prints TSV.
3. `./nathm` → TUI opens; navigate; quit.
4. Create a branch, then `./nathm rename test-branch test-branch-2`.
5. With at least one merged-into-main branch present: `./nathm prune` → review prompt → cancel → re-run with `--yes` → branch deleted.
6. Try `./nathm prune` in a non-git dir → exits 1 with clear error.

- [ ] **Step 5: Commit**

```bash
git add README.md
git commit -m "docs: add README with usage and config"
```

---

## Self-Review Notes

**Spec coverage check (re-run by engineer or reviewer):**

| Spec section | Implemented in task |
|---|---|
| Tech stack (Go + Bubble Tea + Cobra) | Tasks 1, 18 |
| File layout | All tasks |
| Domain model (Branch, Status) | Task 4 |
| Population pipeline (for-each-ref, base detection, ahead/behind) | Tasks 5, 6, 7, 8, 9 |
| TUI list screen (columns, keymap, sort, filter) | Tasks 18, 19, 20 |
| Confirm dialog | Task 21 |
| Subcommands: list / prune / rename / version | Tasks 1, 11, 16, 17 |
| Exit codes (0/1/2) | Tasks 11, 16, 25 |
| Config (TOML, XDG, auto-create) | Task 10 |
| Default protected (current + base) | Task 8 |
| User-configurable protected patterns | Tasks 8, 10 |
| Per-branch failures don't abort batch | Tasks 16, 21 |
| `git fetch --prune` before stale detection | Task 15, 16 |
| TUI panic-safe (deferred recover) | Bubble Tea provides; verified in Task 25 manual test |
| Testing strategy (unit / integration / TUI smoke) | Throughout |

No gaps identified. Type names (`Branch.MergedIntoBase`, `Status.String()`, `LoadConfig.BaseBranches`) are consistent across tasks.
