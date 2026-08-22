package views

import (
	"strings"
	"testing"
)

func TestBuildMinutesModel_ExampleNote(t *testing.T) {
	tree := loadExampleTree(t)
	cfg := loadSeededSettings(t)

	model, err := BuildMinutesModel(tree, cfg)
	if err != nil {
		t.Fatalf("BuildMinutesModel: unexpected error: %v", err)
	}

	if model.Church != churchName {
		t.Errorf("Church = %q, want %q", model.Church, churchName)
	}
	if model.MeetingDate != "August 11, 2026" {
		t.Errorf("MeetingDate = %q, want %q", model.MeetingDate, "August 11, 2026")
	}
	if model.TypeLower != "stated" {
		t.Errorf("TypeLower = %q, want %q", model.TypeLower, "stated")
	}
	if model.Location != "Classroom 7/8" {
		t.Errorf("Location = %q, want %q", model.Location, "Classroom 7/8")
	}
	if model.Time != "5:00 PM" {
		t.Errorf("Time = %q, want %q", model.Time, "5:00 PM")
	}

	wantAbsent := "Dave Murray, Burk Parsons"
	if model.Absent != wantAbsent {
		t.Errorf("Absent = %q, want %q", model.Absent, wantAbsent)
	}

	presentNames := strings.Split(model.Present, ", ")
	if len(presentNames) != 16 {
		t.Fatalf("len(Present names) = %d, want 16: %q", len(presentNames), model.Present)
	}
	for _, name := range presentNames {
		if name == "Dave Murray" || name == "Burk Parsons" {
			t.Errorf("Present includes absent elder %q", name)
		}
	}

	wantMotions := []bool{false, true, true, true, false}
	if len(model.Entries) != len(wantMotions) {
		t.Fatalf("len(Entries) = %d, want %d: %+v", len(model.Entries), len(wantMotions), model.Entries)
	}
	for i, want := range wantMotions {
		if model.Entries[i].Motion != want {
			t.Errorf("Entries[%d].Motion = %v, want %v", i, model.Entries[i].Motion, want)
		}
	}

	bankMotion := string(model.Entries[2].Value)
	if !strings.Contains(bankMotion, "<strong>REMOVE</strong>") {
		t.Errorf("Entries[2].Value missing bolded REMOVE: %q", bankMotion)
	}
	if !strings.Contains(bankMotion, "<strong>ADD</strong>") {
		t.Errorf("Entries[2].Value missing bolded ADD: %q", bankMotion)
	}
}

func TestBuildMinutesModel_MisspelledAbsenceIsHardError(t *testing.T) {
	const body = `# Metadata

## Info

- **date:** 2026-08-11
- **time:** 5:00 PM
- **location:** Classroom 7/8
- **type:** Stated

# Absences

## Not A Real Elder
`
	tree := parseOrFatal(t, body)
	cfg := loadSeededSettings(t)

	_, err := BuildMinutesModel(tree, cfg)
	if err == nil {
		t.Fatal("BuildMinutesModel: got nil error, want a hard error naming the unmatched title")
	}
	if !strings.Contains(err.Error(), "Not A Real Elder") {
		t.Errorf("error = %q, want it to name the unmatched title", err.Error())
	}
}

func TestBuildMinutesModel_NoAbsencesSectionMeansNobodyAbsent(t *testing.T) {
	const body = `# Metadata

## Info

- **date:** 2026-08-11
- **time:** 5:00 PM
- **location:** Classroom 7/8
- **type:** Stated
`
	tree := parseOrFatal(t, body)
	cfg := loadSeededSettings(t)

	model, err := BuildMinutesModel(tree, cfg)
	if err != nil {
		t.Fatalf("BuildMinutesModel: unexpected error: %v", err)
	}
	if model.Absent != "None" {
		t.Errorf("Absent = %q, want %q", model.Absent, "None")
	}
}

func TestRenderMinutes_ExampleNote(t *testing.T) {
	tree := loadExampleTree(t)
	cfg := loadSeededSettings(t)

	var buf strings.Builder
	if err := RenderMinutes(&buf, tree, cfg); err != nil {
		t.Fatalf("RenderMinutes: unexpected error: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"Session Meeting Minutes",
		"Dave Murray, Burk Parsons",
		"<strong>Motion</strong>",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered output missing %q", want)
		}
	}
}
