package branch

import (
	"testing"
	"time"
)

func TestParseForEachRef_Active(t *testing.T) {
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
