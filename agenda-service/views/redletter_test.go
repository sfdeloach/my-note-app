package views

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sfdeloach/my-note-app/agenda-service/settings"
)

func TestBuildRedLetterModel_ExampleNote(t *testing.T) {
	tree := loadExampleTree(t)
	cfg := loadSeededSettings(t)

	model, err := BuildRedLetterModel(tree, cfg)
	if err != nil {
		t.Fatalf("BuildRedLetterModel: unexpected error: %v", err)
	}

	if model.Type != "Stated" {
		t.Errorf("Type = %q, want %q", model.Type, "Stated")
	}
	if model.MeetingDate != "August 11, 2026" {
		t.Errorf("MeetingDate = %q, want %q", model.MeetingDate, "August 11, 2026")
	}
	if model.Time != "5:00 PM" {
		t.Errorf("Time = %q, want %q", model.Time, "5:00 PM")
	}
	if model.Location != "Classroom 7/8" {
		t.Errorf("Location = %q, want %q", model.Location, "Classroom 7/8")
	}

	var absences *RedLetterSection
	for i := range model.Sections {
		if model.Sections[i].Heading == "Absences" {
			absences = &model.Sections[i]
		}
	}
	if absences == nil {
		t.Fatal("Sections missing Absences")
	}

	var burk, dave *RedLetterItem
	for i := range absences.Items {
		switch absences.Items[i].Title {
		case "Burk Parsons":
			burk = &absences.Items[i]
		case "Dave Murray":
			dave = &absences.Items[i]
		}
	}
	if burk == nil || dave == nil {
		t.Fatalf("Absences.Items missing expected names: %+v", absences.Items)
	}
	if len(burk.RedLetter) != 0 {
		t.Errorf("Burk Parsons has %d redLetter notes, want 0 (bare item)", len(burk.RedLetter))
	}
	if len(dave.RedLetter) != 1 {
		t.Fatalf("Dave Murray has %d redLetter notes, want 1", len(dave.RedLetter))
	}
	if !strings.Contains(string(dave.RedLetter[0]), `style="color: firebrick;"`) {
		t.Errorf("Dave Murray redLetter span missing inline color: %q", dave.RedLetter[0])
	}
	if !strings.Contains(string(dave.RedLetter[0]), "(Note: Traveling, notified via email on 8/1)") {
		t.Errorf("Dave Murray redLetter span missing wrapped note text: %q", dave.RedLetter[0])
	}
}

func TestBuildRedLetterModel_EmptySectionOmitted(t *testing.T) {
	tree := parseOrFatal(t, emptySectionFixture)

	model, err := BuildRedLetterModel(tree, settings.Settings{})
	if err != nil {
		t.Fatalf("BuildRedLetterModel: unexpected error: %v", err)
	}
	if len(model.Sections) != 1 || model.Sections[0].Heading != "Notes" {
		t.Errorf("Sections = %+v, want only Notes", model.Sections)
	}
}

// TestRenderRedLetter_CopyContentSurvivesWithoutStyleBlock is the
// automatable half of the roadmap's paste-survival Verify step: the
// #copy-content subtree (what the copy button's Clipboard API call reads)
// must retain the red color via an inline style attribute and must not
// contain a <style> block, since most email clients strip those on paste.
// Actual paste-testing into an email client remains a manual check.
func TestRenderRedLetter_CopyContentSurvivesWithoutStyleBlock(t *testing.T) {
	tree := loadExampleTree(t)
	cfg := loadSeededSettings(t)

	var buf bytes.Buffer
	if err := RenderRedLetter(&buf, tree, cfg); err != nil {
		t.Fatalf("RenderRedLetter: unexpected error: %v", err)
	}
	out := buf.String()

	start := strings.Index(out, `<main id="copy-content">`)
	end := strings.Index(out, "</main>")
	if start == -1 || end == -1 || end < start {
		t.Fatalf("could not locate #copy-content in rendered output:\n%s", out)
	}
	fragment := out[start:end]

	if strings.Contains(fragment, "<style") {
		t.Error("#copy-content fragment contains a <style> block, which most email clients strip on paste")
	}
	if !strings.Contains(fragment, `style="color: firebrick;"`) {
		t.Error("#copy-content fragment missing an inline color style")
	}
}
