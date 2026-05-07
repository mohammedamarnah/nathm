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
