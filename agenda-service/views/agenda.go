package views

import (
	"html/template"
	"io"

	"github.com/sfdeloach/my-note-app/agenda-service/parser"
	"github.com/sfdeloach/my-note-app/agenda-service/settings"
)

// churchName uses the Unicode right single quotation mark (U+2019),
// matching appendix 5's &rsquo; visually.
const churchName = "Saint Andrew’s Chapel"

// agendaFontStack is template.CSS, not string: html/template's contextual
// CSS escaper would otherwise mangle the quotes and commas in a font-family
// value interpolated into a <style> block.
const agendaFontStack template.CSS = `"Times New Roman", Times, serif`

// AgendaSection is one h1 body section on the printed agenda: a heading
// plus its h2 item titles, in document order. No key-values render here.
type AgendaSection struct {
	Heading string
	Ordered bool // true -> <ol>, false -> <ul>
	Items   []string
}

// AgendaModel is the data agenda.tmpl renders.
type AgendaModel struct {
	Church       string
	MeetingDate  string
	ElderColumns int
	Elders       []string // formatted names, active at the meeting date, sorted last-then-first
	Sections     []AgendaSection
	FontStack    template.CSS
}

// BuildAgendaModel builds the printed Agenda's view model from a parsed
// note tree and the loaded settings.
func BuildAgendaModel(tree []*parser.Block, cfg settings.Settings) (AgendaModel, error) {
	m, err := extractMetadata(tree)
	if err != nil {
		return AgendaModel{}, err
	}

	meetingDate, err := FormatDate(m.Date)
	if err != nil {
		return AgendaModel{}, err
	}

	active := settings.ActiveElders(cfg, m.Date)
	elders := make([]string, len(active))
	for i, e := range active {
		elders[i] = e.Name()
	}

	return AgendaModel{
		Church:       churchName,
		MeetingDate:  meetingDate,
		ElderColumns: cfg.ElderColumns,
		Elders:       elders,
		Sections:     buildAgendaSections(tree, cfg),
		FontStack:    agendaFontStack,
	}, nil
}

// buildAgendaSections walks the tree's top-level h1 blocks, skipping
// Metadata and omitting any section with no h2 items entirely — heading and
// all.
func buildAgendaSections(tree []*parser.Block, cfg settings.Settings) []AgendaSection {
	var sections []AgendaSection
	for _, b := range tree {
		if b.Key != "h1" || b.Content == "Metadata" || len(b.Children) == 0 {
			continue
		}
		items := make([]string, len(b.Children))
		for i, item := range b.Children {
			items[i] = item.Content
		}
		sections = append(sections, AgendaSection{
			Heading: b.Content,
			Ordered: cfg.ListTypeFor(b.Content) == settings.Ordered,
			Items:   items,
		})
	}
	return sections
}

// RenderAgenda builds the Agenda view model and executes agenda.tmpl.
func RenderAgenda(w io.Writer, tree []*parser.Block, cfg settings.Settings) error {
	model, err := BuildAgendaModel(tree, cfg)
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, "agenda.tmpl", model)
}
