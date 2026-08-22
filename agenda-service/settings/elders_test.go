package settings

import (
	"strings"
	"testing"
)

func TestElder_Name(t *testing.T) {
	cases := []struct {
		name  string
		elder Elder
		want  string
	}{
		{
			name:  "no suffix",
			elder: Elder{FirstName: "Burk", LastName: "Parsons"},
			want:  "Burk Parsons",
		},
		{
			name:  "with suffix",
			elder: Elder{FirstName: "Steve", LastName: "DeLoach", Suffix: "Jr."},
			want:  "Steve DeLoach Jr.",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.elder.Name()
			if got != c.want {
				t.Errorf("Name() = %q, want %q", got, c.want)
			}
			if strings.Contains(got, ",") {
				t.Errorf("Name() = %q, must never contain a comma", got)
			}
		})
	}
}

func TestElder_IsActiveAt(t *testing.T) {
	removedAt := "2026-03-09"

	cases := []struct {
		name        string
		elder       Elder
		meetingDate string
		want        bool
	}{
		{
			name:        "before activeAt",
			elder:       Elder{ActiveAt: "2020-01-01"},
			meetingDate: "2019-12-31",
			want:        false,
		},
		{
			name:        "exactly on activeAt",
			elder:       Elder{ActiveAt: "2020-01-01"},
			meetingDate: "2020-01-01",
			want:        true,
		},
		{
			name:        "after activeAt, no removedAt",
			elder:       Elder{ActiveAt: "2020-01-01"},
			meetingDate: "2026-08-11",
			want:        true,
		},
		{
			name:        "exactly on removedAt (inclusive boundary)",
			elder:       Elder{ActiveAt: "2014-03-18", RemovedAt: &removedAt},
			meetingDate: "2026-03-09",
			want:        true,
		},
		{
			name:        "one day after removedAt",
			elder:       Elder{ActiveAt: "2014-03-18", RemovedAt: &removedAt},
			meetingDate: "2026-03-10",
			want:        false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.elder.IsActiveAt(c.meetingDate)
			if got != c.want {
				t.Errorf("IsActiveAt(%q) = %v, want %v", c.meetingDate, got, c.want)
			}
		})
	}
}

// TestActiveElders_SeededSettings is the roadmap's headline Stage 3
// Verify assertion, made concrete: 18 active elders at 2026-08-11,
// excluding David Zima and Rey Villavicencio (both past removedAt).
// Coupled to config/settings.json's current roster — update this test if
// the roster is hand-edited later.
func TestActiveElders_SeededSettings(t *testing.T) {
	s, err := Load("../config/settings.json")
	if err != nil {
		t.Fatalf("Load: unexpected error: %v", err)
	}

	active := ActiveElders(s, "2026-08-11")
	if len(active) != 18 {
		names := make([]string, len(active))
		for i, e := range active {
			names[i] = e.Name()
		}
		t.Fatalf("len(ActiveElders) = %d, want 18: %v", len(active), names)
	}

	for _, e := range active {
		if e.LastName == "Zima" {
			t.Error("ActiveElders includes David Zima, who is past removedAt")
		}
		if e.LastName == "Villavicencio" {
			t.Error("ActiveElders includes Rey Villavicencio, who is past removedAt")
		}
	}

	for i := 1; i < len(active); i++ {
		prev, cur := active[i-1], active[i]
		if prev.LastName > cur.LastName ||
			(prev.LastName == cur.LastName && prev.FirstName > cur.FirstName) {
			t.Errorf("ActiveElders not sorted last-then-first at index %d: %q before %q", i, prev.Name(), cur.Name())
		}
	}
}
