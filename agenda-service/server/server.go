// Package server is Agenda Service's HTTP surface: a note listing for the
// two notebooks, and the four rendered views. There is no auth beyond the
// WireGuard tunnel / trusted-LAN boundary the rest of the stack already
// relies on (see CLAUDE.md) — this package adds none of its own.
package server

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/sfdeloach/my-note-app/agenda-service/reader"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

var listTmpl = template.Must(template.ParseFS(templateFS, "templates/*.tmpl"))

// Server implements http.Handler for every route this service exposes.
type Server struct {
	rdr          *reader.Reader
	settingsPath string
	mux          *http.ServeMux
}

// New builds a Server and registers its routes. rdr may be nil only in
// tests exercising request-validation paths that never reach it.
func New(rdr *reader.Reader, settingsPath string) *Server {
	s := &Server{rdr: rdr, settingsPath: settingsPath, mux: http.NewServeMux()}

	s.mux.HandleFunc("GET /{$}", s.handleList(reader.Current, "Session Meetings"))
	s.mux.HandleFunc("GET /archived", s.handleList(reader.Archived, "Archived Session Meetings"))
	s.mux.HandleFunc("GET /view", s.handleView(false))
	s.mux.HandleFunc("GET /download", s.handleView(true))

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
