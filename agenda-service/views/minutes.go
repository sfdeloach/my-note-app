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

// MemberUpdateTable is one rendered membership-change table: a title, the
// header cells its stanza declared, and the data rows. Every cell has
// already been through Bold, so minutes.tmpl emits them unescaped.
type MemberUpdateTable struct {
	Title   string
	Headers []template.HTML
	Rows    [][]template.HTML
}

// memberUpdateTitles is the fixed set of valid Member Updates table
// titles, in render order. Tables always appear in this order regardless
// of the note's stanza order; any other title is a per-note hard error
// (typo protection, same spirit as the Absences name check).
var memberUpdateTitles = []string{"New Members", "Baptisms", "Transfers", "Removals", "Deaths"}

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
	// MemberUpdates is the membership-change tables, in canonical order.
	// Empty (nil) when the note carries no member-updates block or every
	// table in it is empty — the template then renders no section at all.
	MemberUpdates []MemberUpdateTable
	Clerk         string // free text, rendered as-is in the attestation footer
	Moderator     string // free text, rendered as-is in the attestation footer
	FontStack     template.CSS
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

	memberUpdates, err := buildMemberUpdates(tree)
	if err != nil {
		return MinutesModel{}, err
	}

	return MinutesModel{
		Church:        churchName,
		MeetingDate:   meetingDate,
		TypeLower:     strings.ToLower(m.Type),
		Location:      m.Location,
		Time:          m.Time,
		Present:       joinElderNames(present),
		Absent:        joinAbsentNames(absent),
		Entries:       buildMinutesEntries(tree),
		MemberUpdates: memberUpdates,
		Clerk:         m.Clerk,
		Moderator:     m.Moderator,
		FontStack:     minutesActionItemsFontStack,
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

// buildMemberUpdates finds the single top-level "member-updates" fenced
// block (the parser emits it as a depth-0, non-h1 node keyed by its info
// string) and turns each blank-line-separated stanza into a
// MemberUpdateTable. A stanza is: line 1 the title, line 2 the
// pipe-delimited header row, remaining lines the pipe-delimited data rows.
// Stanzas with a title and header but no data rows are dropped; the tables
// come back in memberUpdateTitles order. No block, or a block whose
// stanzas are all empty, yields nil — the template renders no section.
//
// Hard errors, surfaced the same way as absentElders (plain error out of
// BuildMinutesModel → 422): more than one member-updates block, a
// duplicate or unrecognized title, a stanza missing its header row, or a
// data row whose cell count differs from its header.
func buildMemberUpdates(tree []*parser.Block) ([]MemberUpdateTable, error) {
	var block *parser.Block
	for _, b := range tree {
		if b.Key != "member-updates" {
			continue
		}
		if block != nil {
			return nil, fmt.Errorf("views: note has more than one member-updates block")
		}
		block = b
	}
	if block == nil {
		return nil, nil
	}

	byTitle := make(map[string]MemberUpdateTable)
	for _, stanza := range strings.Split(block.Content, "\n\n") {
		lines := strings.Split(strings.Trim(stanza, "\n"), "\n")
		if len(lines) == 1 && strings.TrimSpace(lines[0]) == "" {
			continue // blank run between stanzas
		}

		title := strings.TrimSpace(lines[0])
		if !validMemberUpdateTitle(title) {
			return nil, fmt.Errorf("views: member-updates names table %q, which is not one of %s",
				title, strings.Join(memberUpdateTitles, ", "))
		}
		if _, dup := byTitle[title]; dup {
			return nil, fmt.Errorf("views: member-updates has more than one %q table", title)
		}
		if len(lines) < 2 {
			return nil, fmt.Errorf("views: member-updates table %q has no header row", title)
		}

		headers := splitPipeCells(lines[1])
		var rows [][]template.HTML
		for _, dataLine := range lines[2:] {
			cells := splitPipeCells(dataLine)
			if len(cells) != len(headers) {
				return nil, fmt.Errorf("views: member-updates table %q has a row with %d cells, want %d: %q",
					title, len(cells), len(headers), dataLine)
			}
			row := make([]template.HTML, len(cells))
			for i, c := range cells {
				row[i] = Bold(c)
			}
			rows = append(rows, row)
		}
		if len(rows) == 0 {
			continue // title + header but no data: omit this table
		}

		boldHeaders := make([]template.HTML, len(headers))
		for i, h := range headers {
			boldHeaders[i] = Bold(h)
		}
		byTitle[title] = MemberUpdateTable{Title: title, Headers: boldHeaders, Rows: rows}
	}

	var tables []MemberUpdateTable
	for _, title := range memberUpdateTitles {
		if tbl, ok := byTitle[title]; ok {
			tables = append(tables, tbl)
		}
	}
	return tables, nil
}

func validMemberUpdateTitle(title string) bool {
	for _, t := range memberUpdateTitles {
		if t == title {
			return true
		}
	}
	return false
}

// splitPipeCells splits a pipe-delimited row and trims each cell. There is
// no escape for a literal '|' — that is a deliberate limitation of the
// authoring format, fine for names and dates.
func splitPipeCells(line string) []string {
	cells := strings.Split(line, "|")
	for i := range cells {
		cells[i] = strings.TrimSpace(cells[i])
	}
	return cells
}

// RenderMinutes builds the Minutes view model and executes minutes.tmpl.
func RenderMinutes(w io.Writer, tree []*parser.Block, cfg settings.Settings) error {
	model, err := BuildMinutesModel(tree, cfg)
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, "minutes.tmpl", model)
}
