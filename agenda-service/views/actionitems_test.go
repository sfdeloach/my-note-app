package views

import (
	"strings"
	"testing"
)

func TestBuildActionItemsModel_ExampleNote(t *testing.T) {
	tree := loadExampleTree(t)
	cfg := loadSeededSettings(t)

	model, err := BuildActionItemsModel(tree, cfg)
	if err != nil {
		t.Fatalf("BuildActionItemsModel: unexpected error: %v", err)
	}

	if model.Church != churchName {
		t.Errorf("Church = %q, want %q", model.Church, churchName)
	}
	if model.MeetingDate != "August 11, 2026" {
		t.Errorf("MeetingDate = %q, want %q", model.MeetingDate, "August 11, 2026")
	}

	if len(model.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1: %+v", len(model.Entries), model.Entries)
	}
	entry := model.Entries[0]
	if entry.Title != "Bank Authorization" {
		t.Errorf("Entries[0].Title = %q, want %q", entry.Title, "Bank Authorization")
	}
	if len(entry.Items) != 1 {
		t.Fatalf("len(Entries[0].Items) = %d, want 1: %v", len(entry.Items), entry.Items)
	}
	if !strings.Contains(string(entry.Items[0]), "DeLoach to collect signatures") {
		t.Errorf("Entries[0].Items[0] = %q, want it to contain the DeLoach action item text", entry.Items[0])
	}
}

func TestRenderActionItems_ExampleNote(t *testing.T) {
	tree := loadExampleTree(t)
	cfg := loadSeededSettings(t)

	var buf strings.Builder
	if err := RenderActionItems(&buf, tree, cfg); err != nil {
		t.Fatalf("RenderActionItems: unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "DeLoach to collect signatures") {
		t.Error("rendered output missing the DeLoach action item text")
	}
	if strings.Contains(out, "Appoint a Moderator") {
		t.Error("rendered output includes the title of an item with no actionItem")
	}
}
