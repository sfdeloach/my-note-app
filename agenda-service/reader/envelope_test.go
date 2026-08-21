package reader

import (
	"os"
	"strings"
	"testing"

	"github.com/sfdeloach/my-note-app/agenda-service/parser"
)

// exampleNotePath is the same fixture the parser package tests against,
// reused here so findMetadata's happy path is checked against a real
// authored note rather than a hand-built tree.
const exampleNotePath = "../../docs/agenda-service/appendix/1-example-note.md"

func TestParseTitle(t *testing.T) {
	cases := []struct {
		name      string
		title     string
		wantDate  string
		wantType  string
		wantError bool
	}{
		{
			name:     "well-formed title",
			title:    "2026-08-11 Stated Meeting",
			wantDate: "2026-08-11",
			wantType: "Stated",
		},
		{
			name:     "multi-word type still ends in Meeting",
			title:    "2025-12-02 Called Budget Meeting",
			wantDate: "2025-12-02",
			wantType: "Called Budget",
		},
		{
			// Real note title (2025-11-08) that doesn't end in "Meeting"
			// at all.
			name:      "real title: Called Teleconference",
			title:     "2025-11-08 Called Teleconference",
			wantError: true,
		},
		{
			// Real note title (2025-09-16) that doesn't end in "Meeting"
			// as its final word.
			name:      "real title: Called Meeting/DZ Court",
			title:     "2025-09-16 Called Meeting/DZ Court",
			wantError: true,
		},
		{
			name:      "no date prefix",
			title:     "Stated Meeting",
			wantError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			date, meetingType, err := parseTitle(c.title)
			if c.wantError {
				if err == nil {
					t.Fatalf("parseTitle(%q): got nil error, want one", c.title)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseTitle(%q): unexpected error: %v", c.title, err)
			}
			if date != c.wantDate || meetingType != c.wantType {
				t.Errorf("parseTitle(%q) = %q, %q; want %q, %q", c.title, date, meetingType, c.wantDate, c.wantType)
			}
		})
	}
}

func TestFindMetadata_ExampleNote(t *testing.T) {
	body, err := os.ReadFile(exampleNotePath)
	if err != nil {
		t.Fatalf("reading example note: %v", err)
	}

	blocks, _, err := parser.Parse(string(body))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	date, meetingType, err := findMetadata(blocks)
	if err != nil {
		t.Fatalf("findMetadata returned error: %v", err)
	}
	if date != "2026-08-11" || meetingType != "Stated" {
		t.Errorf("findMetadata = %q, %q; want %q, %q", date, meetingType, "2026-08-11", "Stated")
	}
}

func TestFindMetadata_Malformed(t *testing.T) {
	cases := []struct {
		name   string
		blocks []*parser.Block
	}{
		{
			name:   "no Metadata section",
			blocks: []*parser.Block{{Key: "h1", Content: "Notes"}},
		},
		{
			name: "Metadata with two items",
			blocks: []*parser.Block{{
				Key:     "h1",
				Content: "Metadata",
				Children: []*parser.Block{
					{Key: "item", Content: "Info"},
					{Key: "item", Content: "Extra"},
				},
			}},
		},
		{
			name: "Metadata missing type key",
			blocks: []*parser.Block{{
				Key:     "h1",
				Content: "Metadata",
				Children: []*parser.Block{{
					Key:     "item",
					Content: "Info",
					Children: []*parser.Block{
						{Key: "date", Content: "2026-08-11"},
					},
				}},
			}},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, _, err := findMetadata(c.blocks); err == nil {
				t.Fatalf("findMetadata(%s): got nil error, want one", c.name)
			}
		})
	}
}

func TestCrossCheck(t *testing.T) {
	if err := crossCheck("2026-08-11", "Stated", "2026-08-11", "Stated"); err != nil {
		t.Errorf("crossCheck with matching values: unexpected error: %v", err)
	}

	// Hand-built mismatch case: no real note currently exhibits a
	// title/Metadata drift, so this is tested as pure data rather than by
	// writing a synthetic row into the live database.
	err := crossCheck("2026-08-11", "Stated", "2026-07-14", "Called")
	if err == nil {
		t.Fatal("crossCheck with mismatched values: got nil error, want one")
	}
	for _, want := range []string{"2026-08-11", "Stated", "2026-07-14", "Called"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("crossCheck error %q does not mention %q", err.Error(), want)
		}
	}
}
