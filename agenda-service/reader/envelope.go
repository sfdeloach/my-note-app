package reader

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/sfdeloach/my-note-app/agenda-service/parser"
)

// envelope is the JSON object Joplin stores in items.content for a note
// or notebook row. Only the fields reader actually uses are named here;
// the rest of Joplin's envelope (icon, sharing flags, timestamps, ...) is
// ignored.
type envelope struct {
	Title          string `json:"title"`
	Body           string `json:"body"`
	MarkupLanguage int    `json:"markup_language"`
	DeletedTime    int64  `json:"deleted_time"`
}

// decodeEnvelope JSON-decodes a row's content column. Postgres already
// hands back the bytea as a UTF-8 byte slice, so no separate decode step
// is needed before unmarshaling.
func decodeEnvelope(content []byte) (envelope, error) {
	var e envelope
	if err := json.Unmarshal(content, &e); err != nil {
		return envelope{}, fmt.Errorf("decoding content envelope: %w", err)
	}
	return e, nil
}

var titlePattern = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}) (.+) Meeting$`)

// parseTitle extracts the date and type from a master note title of the
// form "YYYY-MM-DD <Type> Meeting". It errors, naming the title, if the
// pattern doesn't match at all.
func parseTitle(title string) (date, meetingType string, err error) {
	m := titlePattern.FindStringSubmatch(title)
	if m == nil {
		return "", "", fmt.Errorf("title %q does not match the pattern %q", title, "YYYY-MM-DD <Type> Meeting")
	}
	return m[1], m[2], nil
}

// findMetadata locates the "# Metadata" section's date and type values.
// Per the authoring convention, Metadata holds exactly one h2 item (its
// own title is ignored) whose children are the date/time/location/type
// key-value pairs.
func findMetadata(blocks []*parser.Block) (date, meetingType string, err error) {
	var metadata *parser.Block
	for _, b := range blocks {
		if b.Key == "h1" && b.Content == "Metadata" {
			metadata = b
			break
		}
	}
	if metadata == nil {
		return "", "", fmt.Errorf("no Metadata section found")
	}
	if len(metadata.Children) != 1 {
		return "", "", fmt.Errorf("Metadata section has %d items, want exactly 1", len(metadata.Children))
	}

	values := make(map[string]string, len(metadata.Children[0].Children))
	for _, kv := range metadata.Children[0].Children {
		values[kv.Key] = kv.Content
	}

	date, ok := values["date"]
	if !ok {
		return "", "", fmt.Errorf("Metadata section is missing a %q key", "date")
	}
	meetingType, ok = values["type"]
	if !ok {
		return "", "", fmt.Errorf("Metadata section is missing a %q key", "type")
	}
	return date, meetingType, nil
}

// crossCheck compares the date/type parsed from a note's title against
// the date/type recorded in its Metadata section. A meeting note's title
// duplicates these values, and the two can drift when a past note is
// copied forward as a template.
func crossCheck(titleDate, titleType, metaDate, metaType string) error {
	if titleDate != metaDate || titleType != metaType {
		return fmt.Errorf("title says %s/%s, Metadata says %s/%s", titleDate, titleType, metaDate, metaType)
	}
	return nil
}
