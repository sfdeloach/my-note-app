// Package settings loads, validates, and exposes the elder roster and
// per-section list-type configuration that doesn't belong in a note body
// because it changes infrequently. It knows nothing about Joplin,
// Postgres, or the parser's tree — it only understands the settings.json
// shape described in docs/agenda-service/appendix/4-settings-and-data.md.
package settings

import (
	"fmt"
	"os"
)

// ListType is a per-h1-section list rendering type.
type ListType string

const (
	Ordered   ListType = "ordered"
	Unordered ListType = "unordered"
)

// Settings is the fully loaded and validated configuration: the elder
// roster and per-section list types. Unlike reader.Reader, which holds a
// live DB connection, Settings holds no live resource — it is a plain,
// immutable value freshly built by Load on every call, so it is passed by
// value throughout the service.
type Settings struct {
	ListType     map[string]ListType
	ElderColumns int
	Elders       []Elder
}

// Elder is one roster entry.
type Elder struct {
	FirstName string
	LastName  string
	Suffix    string  // "" means none
	ActiveAt  string  // ISO 8601 YYYY-MM-DD
	RemovedAt *string // ISO 8601 YYYY-MM-DD; nil means still active
	// ElderClass is captured but not read by any view — don't wire it up.
	ElderClass string
}

// LoadError describes a failure to load or validate the settings file at
// Path — a bad path, an unreadable file, malformed JSON, or a structural
// validation failure. Settings is a single all-or-nothing file, not a
// collection of independently valid/invalid items, so one error shape is
// enough here.
type LoadError struct {
	Path string
	Err  error
}

func (e *LoadError) Error() string { return fmt.Sprintf("settings: %s: %v", e.Path, e.Err) }
func (e *LoadError) Unwrap() error { return e.Err }

// Load reads and validates the settings file at path. It performs no
// caching and no fallback: every call re-reads and re-parses from disk,
// so a roster edit takes effect without restarting the container. A
// non-nil error is always a *LoadError naming path and the underlying
// failure. Load never panics and never returns a non-zero-value Settings
// alongside a non-nil error.
func Load(path string) (Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Settings{}, &LoadError{Path: path, Err: err}
	}

	s, err := parse(data)
	if err != nil {
		return Settings{}, &LoadError{Path: path, Err: err}
	}

	return s, nil
}

// ListTypeFor returns the list type for the given h1 section title,
// matched exactly and case-sensitively as it appears in the note. A
// title absent from the settings file's listType map defaults to
// Unordered.
func (s Settings) ListTypeFor(sectionTitle string) ListType {
	if lt, ok := s.ListType[sectionTitle]; ok {
		return lt
	}
	return Unordered
}
