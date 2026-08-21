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
	exampleNotePath   = "../../docs/agenda-service/appendix/1-example-note.md"
	scannerResultPath = "../../docs/agenda-service/appendix/3-scanner-result.md"
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
