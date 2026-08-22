package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/sfdeloach/my-note-app/agenda-service/parser"
	"github.com/sfdeloach/my-note-app/agenda-service/settings"
	"github.com/sfdeloach/my-note-app/agenda-service/views"
)

// viewDef describes one of the four renderable views: its query-param
// name, display label, render function, and (for /download) whether it's
// downloadable and what filename prefix to use. Red-Letter is
// screen-only and carries no filename prefix — /download rejects it
// outright rather than leaving FilenamePrefix empty and hoping nothing
// calls it.
type viewDef struct {
	Param          string
	Label          string
	Render         func(w io.Writer, tree []*parser.Block, cfg settings.Settings) error
	Downloadable   bool
	FilenamePrefix string
}

var viewDispatch = []viewDef{
	{Param: "agenda", Label: "Agenda", Render: views.RenderAgenda, Downloadable: true, FilenamePrefix: "session-agenda"},
	{Param: "redletter", Label: "Red-Letter", Render: views.RenderRedLetter, Downloadable: false},
	{Param: "minutes", Label: "Minutes", Render: views.RenderMinutes, Downloadable: true, FilenamePrefix: "session-minutes"},
	{Param: "actionitems", Label: "Action Items", Render: views.RenderActionItems, Downloadable: true, FilenamePrefix: "session-actionitems"},
}

func lookupView(param string) (viewDef, bool) {
	for _, vd := range viewDispatch {
		if vd.Param == param {
			return vd, true
		}
	}
	return viewDef{}, false
}

// handleView serves GET /view (download=false) and GET /download
// (download=true) — the only differences are the Content-Disposition
// header and that download rejects Red-Letter, which has nothing to
// download.
func (s *Server) handleView(download bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		viewParam := r.URL.Query().Get("view")
		if id == "" || viewParam == "" {
			writeBadRequest(w, "both id and view query parameters are required")
			return
		}

		vd, ok := lookupView(viewParam)
		if !ok {
			writeBadRequest(w, fmt.Sprintf("unknown view %q", viewParam))
			return
		}
		if download && !vd.Downloadable {
			writeBadRequest(w, fmt.Sprintf("%s has no download", vd.Label))
			return
		}

		nb, err := s.rdr.GetNote(r.Context(), id)
		if err != nil {
			writeNoteError(w, id, err)
			return
		}

		cfg, err := settings.Load(s.settingsPath)
		if err != nil {
			writeInternal(w, "loading settings", err)
			return
		}

		// GetNote only succeeds once evaluateNote has already confirmed
		// the body parses cleanly, so a failure here would mean an
		// internal inconsistency, not a normal per-note content problem.
		tree, _, err := parser.Parse(nb.Body)
		if err != nil {
			writeInternal(w, fmt.Sprintf("re-parsing note %q", nb.Title), err)
			return
		}

		// Render into a buffer first: if this fails partway, the response
		// must still become a clean error, not a half-written 200.
		var buf bytes.Buffer
		if err := vd.Render(&buf, tree, cfg); err != nil {
			writeNoteProblem(w, nb.Title, err)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if download {
			filename := fmt.Sprintf("%s-%s.html", vd.FilenamePrefix, nb.Date)
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
		}
		if _, err := buf.WriteTo(w); err != nil {
			log.Printf("server: writing view response: %v", err)
		}
	}
}
