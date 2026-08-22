package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/sfdeloach/my-note-app/agenda-service/reader"
)

// writeBadRequest responds with msg verbatim. Only for messages already
// safe to show a client: a missing/unknown query param, never anything
// that might carry DB or filesystem detail.
func writeBadRequest(w http.ResponseWriter, msg string) {
	http.Error(w, msg, http.StatusBadRequest)
}

// writeNoteProblem responds with a per-note problem — a note that exists
// and was fetched, but whose content this service can't render (a
// malformed body, or a view-specific hard error like Minutes' unmatched
// absentee name). The message is safe to show: it's always one of this
// service's own generated errors, naming the note and the problem, never
// a raw DB error.
func writeNoteProblem(w http.ResponseWriter, title string, err error) {
	http.Error(w, fmt.Sprintf("note %q: %v", title, err), http.StatusUnprocessableEntity)
}

// writeNoteError maps a GetNote error to a clean response: not-found and
// per-note problems get a specific, safe message; anything else (a DB or
// infra failure) is logged server-side and answered generically.
func writeNoteError(w http.ResponseWriter, id string, err error) {
	if errors.Is(err, reader.ErrNoteNotFound) {
		http.Error(w, "note not found", http.StatusNotFound)
		return
	}

	var noteErr *reader.NoteError
	if errors.As(err, &noteErr) {
		http.Error(w, fmt.Sprintf("note %s: %v", noteErr.ID, noteErr.Err), http.StatusUnprocessableEntity)
		return
	}

	writeInternal(w, "fetching note "+id, err)
}

// writeInternal logs the real error server-side and responds with a
// generic message — used for anything that might carry DB/infra detail
// that must never reach the client.
func writeInternal(w http.ResponseWriter, action string, err error) {
	log.Printf("server: %s: %v", action, err)
	http.Error(w, "internal error", http.StatusInternalServerError)
}
