package server

import (
	"bytes"
	"log"
	"net/http"
	"net/url"

	"github.com/sfdeloach/my-note-app/agenda-service/reader"
)

// listPageData is what list.tmpl renders for both / and /archived.
type listPageData struct {
	Title string
	Notes []listNote
}

// listNote is one row of the listing. Views is empty when Err is set — a
// note with a per-note problem shows the problem instead of dead links.
type listNote struct {
	Title string
	Date  string
	Type  string
	Err   string
	Views []noteView
}

// noteView is one view's link (and, if downloadable, download link) for a
// single note, precomputed so the template only does substitution, never
// URL construction.
type noteView struct {
	Label       string
	ViewURL     string
	DownloadURL string // "" when this view has no download
}

func (s *Server) handleList(scope reader.Scope, title string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		notes, err := s.rdr.ListNotes(r.Context(), scope)
		if err != nil {
			writeInternal(w, "listing notes", err)
			return
		}

		data := listPageData{Title: title, Notes: make([]listNote, len(notes))}
		for i, n := range notes {
			ln := listNote{Title: n.Title, Date: n.Date, Type: n.Type}
			if n.Err != nil {
				ln.Err = n.Err.Error()
			} else {
				ln.Views = noteViews(n.ID)
			}
			data.Notes[i] = ln
		}

		var buf bytes.Buffer
		if err := listTmpl.ExecuteTemplate(&buf, "list.tmpl", data); err != nil {
			writeInternal(w, "rendering list", err)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := buf.WriteTo(w); err != nil {
			log.Printf("server: writing list response: %v", err)
		}
	}
}

// noteViews builds every view's link for one note, in viewDispatch's
// display order.
func noteViews(id string) []noteView {
	q := "id=" + url.QueryEscape(id) + "&view="
	nvs := make([]noteView, len(viewDispatch))
	for i, vd := range viewDispatch {
		nv := noteView{Label: vd.Label, ViewURL: "/view?" + q + vd.Param}
		if vd.Downloadable {
			nv.DownloadURL = "/download?" + q + vd.Param
		}
		nvs[i] = nv
	}
	return nvs
}
