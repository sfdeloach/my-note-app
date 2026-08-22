package views

import (
	"html/template"
	"io"

	"github.com/sfdeloach/my-note-app/agenda-service/parser"
	"github.com/sfdeloach/my-note-app/agenda-service/settings"
)

// RedLetterItem is one h2 item on the Red-Letter Agenda: its title, plus
// zero or more redLetter notes (keys can repeat) in document order. An item
// with none renders as a bare list item.
type RedLetterItem struct {
	Title     string
	RedLetter []template.HTML
}

// RedLetterSection is one h1 body section.
type RedLetterSection struct {
	Heading string
	Ordered bool
	Items   []RedLetterItem
}

// RedLetterModel is the data redletter.tmpl renders. Unlike Agenda, there
// is no church name or roll call — see appendix 6.
type RedLetterModel struct {
	Type        string // Metadata type, cased as authored
	MeetingDate string
	Time        string
	Location    string
	Sections    []RedLetterSection
}

// BuildRedLetterModel builds the Red-Letter Agenda's view model from a
// parsed note tree and the loaded settings.
func BuildRedLetterModel(tree []*parser.Block, cfg settings.Settings) (RedLetterModel, error) {
	m, err := extractMetadata(tree)
	if err != nil {
		return RedLetterModel{}, err
	}

	meetingDate, err := FormatDate(m.Date)
	if err != nil {
		return RedLetterModel{}, err
	}

	return RedLetterModel{
		Type:        m.Type,
		MeetingDate: meetingDate,
		Time:        m.Time,
		Location:    m.Location,
		Sections:    buildRedLetterSections(tree, cfg),
	}, nil
}

// buildRedLetterSections walks the tree the same way buildAgendaSections
// does (skip Metadata, omit empty sections) but each item also collects its
// redLetter children. Kept separate from buildAgendaSections rather than
// factored into one generic walk: the two produce different item shapes,
// and the walk itself is small enough that a shared generic helper would
// cost more than it saves for two call sites.
func buildRedLetterSections(tree []*parser.Block, cfg settings.Settings) []RedLetterSection {
	var sections []RedLetterSection
	for _, b := range tree {
		if b.Key != "h1" || b.Content == "Metadata" || len(b.Children) == 0 {
			continue
		}
		items := make([]RedLetterItem, len(b.Children))
		for i, item := range b.Children {
			var notes []template.HTML
			for _, kv := range item.Children {
				if kv.Key != "redLetter" {
					continue
				}
				notes = append(notes, redLetterSpan(kv.Content))
			}
			items[i] = RedLetterItem{Title: item.Content, RedLetter: notes}
		}
		sections = append(sections, RedLetterSection{
			Heading: b.Content,
			Ordered: cfg.ListTypeFor(b.Content) == settings.Ordered,
			Items:   items,
		})
	}
	return sections
}

// redLetterSpan wraps a redLetter value as the literal "(Note: …)" span
// with the red color inlined on the style attribute — most email clients
// strip <style> blocks on paste, so the inline attribute is what survives.
// The class is kept too, for the on-page view.
func redLetterSpan(value string) template.HTML {
	return template.HTML(`<span class="red-letter" style="color: firebrick;">(Note: ` + string(Bold(value)) + `)</span>`)
}

// RenderRedLetter builds the Red-Letter Agenda view model and executes
// redletter.tmpl.
func RenderRedLetter(w io.Writer, tree []*parser.Block, cfg settings.Settings) error {
	model, err := BuildRedLetterModel(tree, cfg)
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, "redletter.tmpl", model)
}
