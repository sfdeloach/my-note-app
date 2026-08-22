package settings

import (
	"errors"
	"testing"
)

// TestLoad_SeededFile confirms config/settings.json (seeded verbatim from
// docs/agenda-service/appendix/4-settings-and-data.md) loads cleanly —
// the roadmap's first Stage 3 Verify bullet.
func TestLoad_SeededFile(t *testing.T) {
	s, err := Load("../config/settings.json")
	if err != nil {
		t.Fatalf("Load: unexpected error: %v", err)
	}
	if s.ElderColumns != 3 {
		t.Errorf("ElderColumns = %d, want 3", s.ElderColumns)
	}
	if len(s.Elders) != 20 {
		t.Errorf("len(Elders) = %d, want 20", len(s.Elders))
	}

	var deloach *Elder
	for i := range s.Elders {
		if s.Elders[i].LastName == "DeLoach" {
			deloach = &s.Elders[i]
			break
		}
	}
	if deloach == nil {
		t.Fatal("Steve DeLoach not found in seeded roster")
	}
	if deloach.Suffix != "Jr." {
		t.Errorf("DeLoach.Suffix = %q, want %q", deloach.Suffix, "Jr.")
	}
}

// TestLoad_Malformed confirms the roadmap's second Stage 3 Verify bullet:
// a deliberately malformed settings file produces a clear error naming
// the file and the underlying JSON error, not a crash or a partial
// success.
func TestLoad_Malformed(t *testing.T) {
	s, err := Load("testdata/malformed.json")
	if err == nil {
		t.Fatal("Load: expected error for malformed JSON, got nil")
	}

	var loadErr *LoadError
	if !errors.As(err, &loadErr) {
		t.Fatalf("Load: error is not a *LoadError: %v", err)
	}
	if loadErr.Path != "testdata/malformed.json" {
		t.Errorf("LoadError.Path = %q, want %q", loadErr.Path, "testdata/malformed.json")
	}
	if loadErr.Unwrap() == nil {
		t.Error("LoadError.Unwrap() = nil, want the underlying JSON error")
	}
	if s.ListType != nil || s.ElderColumns != 0 || s.Elders != nil {
		t.Errorf("Load: returned non-zero Settings alongside an error: %+v", s)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("testdata/does-not-exist.json")
	if err == nil {
		t.Fatal("Load: expected error for missing file, got nil")
	}

	var loadErr *LoadError
	if !errors.As(err, &loadErr) {
		t.Fatalf("Load: error is not a *LoadError: %v", err)
	}
	if loadErr.Path != "testdata/does-not-exist.json" {
		t.Errorf("LoadError.Path = %q, want %q", loadErr.Path, "testdata/does-not-exist.json")
	}
	if loadErr.Unwrap() == nil {
		t.Error("LoadError.Unwrap() = nil, want the underlying os.ReadFile error")
	}
}

func TestListTypeFor(t *testing.T) {
	s := Settings{ListType: map[string]ListType{"Absences": Ordered}}

	if got := s.ListTypeFor("Absences"); got != Ordered {
		t.Errorf("ListTypeFor(Absences) = %q, want %q", got, Ordered)
	}
	if got := s.ListTypeFor("Nonexistent Section"); got != Unordered {
		t.Errorf("ListTypeFor(Nonexistent Section) = %q, want %q (default)", got, Unordered)
	}
}
