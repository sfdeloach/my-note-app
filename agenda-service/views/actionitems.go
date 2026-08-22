package views

import (
	"html/template"
	"io"

	"github.com/sfdeloach/my-note-app/agenda-service/parser"
	"github.com/sfdeloach/my-note-app/agenda-service/settings"
)

// ActionItemEntry is one h2 item carrying at least one actionItem child:
// its title, plus each actionItem value in document order.
type ActionItemEntry struct {
	Title string
	Items []template.HTML
}

// ActionItemsModel is the data actionitems.tmpl renders. Metadata's type
// is not used, per the brief.
type ActionItemsModel struct {
	Church      string
	MeetingDate string
	Entries     []ActionItemEntry
	FontStack   template.CSS
}

// BuildActionItemsModel builds the Action Items view model from a parsed
// note tree and the loaded settings.
func BuildActionItemsModel(tree []*parser.Block, cfg settings.Settings) (ActionItemsModel, error) {
	m, err := extractMetadata(tree)
	if err != nil {
		return ActionItemsModel{}, err
	}

	meetingDate, err := FormatDate(m.Date)
	if err != nil {
		return ActionItemsModel{}, err
	}

	return ActionItemsModel{
		Church:      churchName,
		MeetingDate: meetingDate,
		Entries:     buildActionItemEntries(tree),
		FontStack:   minutesActionItemsFontStack,
	}, nil
}

// buildActionItemEntries walks the tree's top-level h1 blocks (skipping
// Metadata) and, for every h2 item, collects its actionItem children in
// document order. An item with none is omitted entirely, and so is any
// section heading.
func buildActionItemEntries(tree []*parser.Block) []ActionItemEntry {
	var entries []ActionItemEntry
	for _, b := range tree {
		if b.Key != "h1" || b.Content == "Metadata" {
			continue
		}
		for _, item := range b.Children {
			var actionItems []template.HTML
			for _, kv := range item.Children {
				if kv.Key != "actionItem" {
					continue
				}
				actionItems = append(actionItems, Bold(kv.Content))
			}
			if len(actionItems) == 0 {
				continue
			}
			entries = append(entries, ActionItemEntry{Title: item.Content, Items: actionItems})
		}
	}
	return entries
}

// RenderActionItems builds the Action Items view model and executes
// actionitems.tmpl.
func RenderActionItems(w io.Writer, tree []*parser.Block, cfg settings.Settings) error {
	model, err := BuildActionItemsModel(tree, cfg)
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, "actionitems.tmpl", model)
}
