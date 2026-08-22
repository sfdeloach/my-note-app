package views

import (
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/sfdeloach/my-note-app/agenda-service/parser"
	"github.com/sfdeloach/my-note-app/agenda-service/settings"
)

// minutesActionItemsFontStack is the shared Cambria stack for Minutes and
// Action Items, deliberately not unified with Agenda's Times New Roman
// (see print.tmpl's doc comment).
const minutesActionItemsFontStack template.CSS = `Cambria, Cochin, Georgia, Times, "Times New Roman", serif`

// MinutesEntry is one motion or comments value, in document order. Motion
// carries the "Motion" label; comments doesn't.
type MinutesEntry struct {
	Motion bool
	Value  template.HTML
}

// MinutesModel is the data minutes.tmpl renders.
type MinutesModel struct {
	Church      string
	MeetingDate string
	TypeLower   string
	Location    string
	Time        string
	Present     string // comma-joined names, last-then-first
	Absent      string // comma-joined names, last-then-first, or "None"
	Entries     []MinutesEntry
	FontStack   template.CSS
}

// BuildMinutesModel builds the Minutes view model from a parsed note tree
// and the loaded settings.
func BuildMinutesModel(tree []*parser.Block, cfg settings.Settings) (MinutesModel, error) {
	m, err := extractMetadata(tree)
	if err != nil {
		return MinutesModel{}, err
	}

	meetingDate, err := FormatDate(m.Date)
	if err != nil {
		return MinutesModel{}, err
	}

	absent, err := absentElders(tree, cfg)
	if err != nil {
		return MinutesModel{}, err
	}
	present := presentElders(cfg, m.Date, absent)

	return MinutesModel{
		Church:      churchName,
		MeetingDate: meetingDate,
		TypeLower:   strings.ToLower(m.Type),
		Location:    m.Location,
		Time:        m.Time,
		Present:     joinElderNames(present),
		Absent:      joinAbsentNames(absent),
		Entries:     buildMinutesEntries(tree),
		FontStack:   minutesActionItemsFontStack,
	}, nil
}

// absentElders finds the "# Absences" section (its absence is not an
// error — nobody absent) and matches each h2 title case-insensitively
// against "{FirstName} {LastName}" for every elder in cfg (suffix excluded
// from the match key). A title matching no elder is a hard error, per the
// brief: a misspelled name under Absences must never produce a silently
// wrong roster.
func absentElders(tree []*parser.Block, cfg settings.Settings) ([]settings.Elder, error) {
	var absences *parser.Block
	for _, b := range tree {
		if b.Key == "h1" && b.Content == "Absences" {
			absences = b
			break
		}
	}
	if absences == nil {
		return nil, nil
	}

	var absent []settings.Elder
	for _, item := range absences.Children {
		matched := false
		for _, e := range cfg.Elders {
			if strings.EqualFold(item.Content, e.FirstName+" "+e.LastName) {
				absent = append(absent, e)
				matched = true
				break
			}
		}
		if !matched {
			return nil, fmt.Errorf("views: Absences names %q, which matches no elder in settings", item.Content)
		}
	}
	settings.SortElders(absent)
	return absent, nil
}

// presentElders is the elders active at meetingDate minus those in absent,
// excluded by first+last name rather than struct equality — Elder holds a
// *string field, so independently-built copies of the same elder aren't
// == comparable.
func presentElders(cfg settings.Settings, meetingDate string, absent []settings.Elder) []settings.Elder {
	excluded := make(map[string]bool, len(absent))
	for _, e := range absent {
		excluded[e.FirstName+"\x00"+e.LastName] = true
	}

	var present []settings.Elder
	for _, e := range settings.ActiveElders(cfg, meetingDate) {
		if !excluded[e.FirstName+"\x00"+e.LastName] {
			present = append(present, e)
		}
	}
	return present
}

func joinElderNames(elders []settings.Elder) string {
	names := make([]string, len(elders))
	for i, e := range elders {
		names[i] = e.Name()
	}
	return strings.Join(names, ", ")
}

func joinAbsentNames(absent []settings.Elder) string {
	if len(absent) == 0 {
		return "None"
	}
	return joinElderNames(absent)
}

// buildMinutesEntries walks the tree's top-level h1 blocks (skipping
// Metadata) and, for every h2 item's children in document order, collects
// motion and comments values. Item titles and section headings never
// surface — that's what makes it legitimate for the adjournment motion and
// closing prayer to live under the "Bank Authorization" item in the
// master note. Kept separate from buildAgendaSections/
// buildRedLetterSections: this walk filters by descendant key rather than
// by item presence, and produces a flat list with no heading/title at all,
// so forcing it through the existing helpers would fight their shape.
func buildMinutesEntries(tree []*parser.Block) []MinutesEntry {
	var entries []MinutesEntry
	for _, b := range tree {
		if b.Key != "h1" || b.Content == "Metadata" {
			continue
		}
		for _, item := range b.Children {
			for _, kv := range item.Children {
				switch kv.Key {
				case "motion":
					entries = append(entries, MinutesEntry{Motion: true, Value: Bold(kv.Content)})
				case "comments":
					entries = append(entries, MinutesEntry{Motion: false, Value: Bold(kv.Content)})
				}
			}
		}
	}
	return entries
}

// RenderMinutes builds the Minutes view model and executes minutes.tmpl.
func RenderMinutes(w io.Writer, tree []*parser.Block, cfg settings.Settings) error {
	model, err := BuildMinutesModel(tree, cfg)
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, "minutes.tmpl", model)
}
