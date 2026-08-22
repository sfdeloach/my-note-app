// Package views renders the four derived views (Agenda, Red-Letter Agenda,
// Minutes, Action Items) from a parsed note tree plus settings.Settings.
// Views are pure functions of their inputs — no Joplin or Postgres
// knowledge, matching the brief's func(tree []*parser.Block, cfg
// settings.Settings) (ViewModel, error) shape.
package views

import (
	"html/template"
	"regexp"
)

var boldPattern = regexp.MustCompile(`\*\*(.+?)\*\*`)

// Bold HTML-escapes s, then substitutes paired "**…**" delimiters into
// <strong>…</strong>. Escaping first is what keeps an author from
// injecting markup through a value; substituting second is safe because
// HTML-escaping never touches '*'. An unmatched "**" (no closing pair) is
// left as literal asterisks, per the brief — visible in the output is
// signal enough.
func Bold(s string) template.HTML {
	escaped := template.HTMLEscapeString(s)
	return template.HTML(boldPattern.ReplaceAllString(escaped, "<strong>$1</strong>"))
}
