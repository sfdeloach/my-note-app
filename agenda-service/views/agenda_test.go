package views

import (
	"os"
	"testing"

	"github.com/sfdeloach/my-note-app/agenda-service/parser"
	"github.com/sfdeloach/my-note-app/agenda-service/settings"
)

// These fixtures live under docs/agenda-service/ rather than being
// duplicated here, so the views tests can never drift from the brief's
// documented example — same rationale as parser_test.go.
const (
	exampleNotePath    = "../../docs/agenda-service/appendix/1-example-note.md"
	seededSettingsPath = "../config/settings.json"
)

func parseOrFatal(t *testing.T, body string) []*parser.Block {
	t.Helper()
	blocks, _, err := parser.Parse(body)
	if err != nil {
		t.Fatalf("Parse: unexpected error: %v", err)
	}
	return blocks
}

func loadExampleTree(t *testing.T) []*parser.Block {
	t.Helper()
	body, err := os.ReadFile(exampleNotePath)
	if err != nil {
		t.Fatalf("reading example note: %v", err)
	}
	return parseOrFatal(t, string(body))
}

func loadSeededSettings(t *testing.T) settings.Settings {
	t.Helper()
	s, err := settings.Load(seededSettingsPath)
	if err != nil {
		t.Fatalf("settings.Load: unexpected error: %v", err)
	}
	return s
}

// emptySectionFixture has one non-empty body section ("Notes") and one h1
// with zero h2 children ("Empty Section"), to exercise the "empty sections
// are omitted entirely" rule — the real example note has no empty section.
const emptySectionFixture = `# Metadata

## Info

- **date:** 2026-08-11
- **time:** 5:00 PM
- **location:** Classroom 7/8
- **type:** Stated

# Notes

## Something

# Empty Section
`

func TestBuildAgendaModel_ExampleNote(t *testing.T) {
	tree := loadExampleTree(t)
	cfg := loadSeededSettings(t)

	model, err := BuildAgendaModel(tree, cfg)
	if err != nil {
		t.Fatalf("BuildAgendaModel: unexpected error: %v", err)
	}

	if model.Church != churchName {
		t.Errorf("Church = %q, want %q", model.Church, churchName)
	}
	if model.MeetingDate != "August 11, 2026" {
		t.Errorf("MeetingDate = %q, want %q", model.MeetingDate, "August 11, 2026")
	}
	if model.ElderColumns != 3 {
		t.Errorf("ElderColumns = %d, want 3", model.ElderColumns)
	}
	if len(model.Elders) != 18 {
		t.Fatalf("len(Elders) = %d, want 18: %v", len(model.Elders), model.Elders)
	}
	for _, name := range model.Elders {
		if name == "David Zima" || name == "Rey Villavicencio" {
			t.Errorf("Elders includes %q, who is past removedAt", name)
		}
	}

	var absences *AgendaSection
	for i := range model.Sections {
		if model.Sections[i].Heading == "Metadata" {
			t.Error("Sections includes Metadata, which must never render as a body section")
		}
		if model.Sections[i].Heading == "Absences" {
			absences = &model.Sections[i]
		}
	}
	if absences == nil {
		t.Fatal("Sections missing Absences")
	}
	if !absences.Ordered {
		t.Error("Absences.Ordered = false, want true")
	}
	wantItems := []string{"Burk Parsons", "Dave Murray"}
	if len(absences.Items) != len(wantItems) {
		t.Fatalf("Absences.Items = %v, want %v", absences.Items, wantItems)
	}
	for i, item := range absences.Items {
		if item != wantItems[i] {
			t.Errorf("Absences.Items[%d] = %q, want %q", i, item, wantItems[i])
		}
	}
}

func TestBuildAgendaModel_EmptySectionOmitted(t *testing.T) {
	tree := parseOrFatal(t, emptySectionFixture)

	model, err := BuildAgendaModel(tree, settings.Settings{ElderColumns: 1})
	if err != nil {
		t.Fatalf("BuildAgendaModel: unexpected error: %v", err)
	}

	if len(model.Sections) != 1 {
		t.Fatalf("len(Sections) = %d, want 1 (only Notes): %+v", len(model.Sections), model.Sections)
	}
	if model.Sections[0].Heading != "Notes" {
		t.Errorf("Sections[0].Heading = %q, want %q", model.Sections[0].Heading, "Notes")
	}
}
