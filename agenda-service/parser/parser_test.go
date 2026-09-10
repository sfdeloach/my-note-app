package parser

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

// These fixtures live in docs/agenda-service/appendix/ rather than being
// duplicated here, so the parser's test can never drift from the brief's
// documented example.
const (
	exampleNotePath   = "../../docs/agenda-service/appendix/01-example-note.md"
	scannerResultPath = "../../docs/agenda-service/appendix/03-scanner-result.md"
)

func TestParseExampleNote(t *testing.T) {
	body, err := os.ReadFile(exampleNotePath)
	if err != nil {
		t.Fatalf("reading example note: %v", err)
	}

	blocks, warnings, err := Parse(string(body))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	want := loadExpectedTree(t)
	if !reflect.DeepEqual(blocks, want) {
		gotJSON, _ := json.MarshalIndent(blocks, "", "  ")
		wantJSON, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("parsed tree does not match fixture.\ngot:\n%s\n\nwant:\n%s", gotJSON, wantJSON)
	}

	// Per the fixture's own note: assert on key names and warning count,
	// not exact line numbers, so the fixture doesn't break every time the
	// example note is edited.
	wantKeys := []string{"motion", "comments", "actionItem"}
	if len(warnings) != len(wantKeys) {
		t.Fatalf("got %d warnings, want %d: %+v", len(warnings), len(wantKeys), warnings)
	}
	for i, w := range warnings {
		if w.Key != wantKeys[i] {
			t.Errorf("warning %d: got key %q, want %q", i, w.Key, wantKeys[i])
		}
	}
}

func TestParseFencedBlockTerminated(t *testing.T) {
	body := "# Member Updates\n\n" +
		"```member-updates\n" +
		"New Members\n" +
		"First | Last\n" +
		"Jane | Doe\n" +
		"```\n"

	blocks, warnings, err := Parse(body)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("got %d warnings, want 0: %+v", len(warnings), warnings)
	}
	if len(blocks) != 2 {
		t.Fatalf("got %d top-level blocks, want 2: %+v", len(blocks), blocks)
	}

	if blocks[0].Key != "h1" || blocks[0].Content != "Member Updates" || blocks[0].Children != nil {
		t.Errorf("blocks[0] = %+v, want an empty h1 %q", blocks[0], "Member Updates")
	}

	fence := blocks[1]
	if fence.Key != "member-updates" {
		t.Errorf("fence key = %q, want %q", fence.Key, "member-updates")
	}
	if fence.Children != nil {
		t.Errorf("fence children = %+v, want nil", fence.Children)
	}
	wantContent := "New Members\nFirst | Last\nJane | Doe"
	if fence.Content != wantContent {
		t.Errorf("fence content = %q, want %q", fence.Content, wantContent)
	}
}

func TestParseFencedBlockCapturesOtherwiseUnrecognizedLines(t *testing.T) {
	// A markdown table row and a bare prose line both fail the parser
	// outside a fence; inside one they must be taken verbatim.
	body := "```member-updates\n" +
		"| a | b |\n" +
		"random text with no leading dash\n" +
		"```\n"

	blocks, _, err := Parse(body)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(blocks) != 1 {
		t.Fatalf("got %d blocks, want 1", len(blocks))
	}
	wantContent := "| a | b |\nrandom text with no leading dash"
	if blocks[0].Content != wantContent {
		t.Errorf("content = %q, want %q", blocks[0].Content, wantContent)
	}
}

func TestParseFencedBlockInfoString(t *testing.T) {
	cases := []struct {
		name    string
		open    string
		wantKey string
	}{
		{"named", "```json", "json"},
		{"empty", "```", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := tc.open + "\nfoo\n```\n"
			blocks, _, err := Parse(body)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			if len(blocks) != 1 {
				t.Fatalf("got %d blocks, want 1", len(blocks))
			}
			if blocks[0].Key != tc.wantKey {
				t.Errorf("key = %q, want %q", blocks[0].Key, tc.wantKey)
			}
			if blocks[0].Content != "foo" {
				t.Errorf("content = %q, want %q", blocks[0].Content, "foo")
			}
		})
	}
}

func TestParseFencedBlockUnterminated(t *testing.T) {
	body := "# Notes\n\n```member-updates\nNew Members\nJane | Doe\n"

	blocks, warnings, err := Parse(body)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("got %d warnings, want 1: %+v", len(warnings), warnings)
	}
	if warnings[0].Line != 3 {
		t.Errorf("warning line = %d, want 3 (the fence-open line)", warnings[0].Line)
	}
	if len(blocks) != 2 || blocks[1].Key != "member-updates" {
		t.Fatalf("want an h1 then a member-updates block, got %+v", blocks)
	}
	if blocks[1].Content != "New Members\nJane | Doe" {
		t.Errorf("content = %q, want the captured-to-EOF body", blocks[1].Content)
	}
}

func TestParseFencedBlockNotNestedUnderHeading(t *testing.T) {
	body := "# Member Updates\n\n```member-updates\nDeaths\nName | Date\n```\n"

	blocks, _, err := Parse(body)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if len(blocks) != 2 {
		t.Fatalf("got %d top-level blocks, want 2 siblings", len(blocks))
	}
	if blocks[0].Children != nil {
		t.Errorf("the h1 has children %+v, want the fence to be a sibling, not a child", blocks[0].Children)
	}
}

// loadExpectedTree extracts the fenced ```json code block from the scanner
// result fixture and unmarshals it into the tree shape Parse produces.
func loadExpectedTree(t *testing.T) []*Block {
	t.Helper()

	raw, err := os.ReadFile(scannerResultPath)
	if err != nil {
		t.Fatalf("reading scanner result fixture: %v", err)
	}

	const fence = "```json"
	content := string(raw)
	start := strings.Index(content, fence)
	if start == -1 {
		t.Fatalf("no %s fence found in %s", fence, scannerResultPath)
	}
	start += len(fence)

	end := strings.Index(content[start:], "```")
	if end == -1 {
		t.Fatalf("unterminated %s fence in %s", fence, scannerResultPath)
	}

	var expected []*Block
	if err := json.Unmarshal([]byte(content[start:start+end]), &expected); err != nil {
		t.Fatalf("unmarshaling expected tree: %v", err)
	}
	return expected
}
