package views

import (
	"fmt"

	"github.com/sfdeloach/my-note-app/agenda-service/parser"
)

// meta holds the six key-value pairs from a note's "# Metadata" section.
type meta struct {
	Date, Time, Location, Type, Clerk, Moderator string
}

// extractMetadata locates the "# Metadata" section and reads its
// date/time/location/type/clerk/moderator values. Metadata holds exactly
// one h2 (its own title is ignored, per the authoring convention)
// containing the six key-values. Views need this independently of reader —
// the brief's view signature takes only the parsed tree and settings, not
// the already title-validated reader.Note, so the values used for head
// matter must come from the tree itself. clerk/moderator are free text,
// consumed only by the Minutes view's attestation footer.
func extractMetadata(blocks []*parser.Block) (meta, error) {
	var metadata *parser.Block
	for _, b := range blocks {
		if b.Key == "h1" && b.Content == "Metadata" {
			metadata = b
			break
		}
	}
	if metadata == nil {
		return meta{}, fmt.Errorf("views: no Metadata section found")
	}
	if len(metadata.Children) != 1 {
		return meta{}, fmt.Errorf("views: Metadata section has %d items, want exactly 1", len(metadata.Children))
	}

	values := make(map[string]string, len(metadata.Children[0].Children))
	for _, kv := range metadata.Children[0].Children {
		values[kv.Key] = kv.Content
	}

	m := meta{
		Date:      values["date"],
		Time:      values["time"],
		Location:  values["location"],
		Type:      values["type"],
		Clerk:     values["clerk"],
		Moderator: values["moderator"],
	}

	for _, kv := range []struct{ name, val string }{
		{"date", m.Date}, {"time", m.Time}, {"location", m.Location}, {"type", m.Type},
		{"clerk", m.Clerk}, {"moderator", m.Moderator},
	} {
		if kv.val == "" {
			return meta{}, fmt.Errorf("views: Metadata section is missing a %q key", kv.name)
		}
	}

	return m, nil
}
