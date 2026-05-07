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
//
//	""                          → no upstream
//	"[]"                        → up to date
//	"[gone]"                    → upstream deleted
//	"[ahead N]"
//	"[behind N]"
//	"[ahead N, behind M]"
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
