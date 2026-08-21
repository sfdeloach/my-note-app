// Package reader is the only part of Agenda Service that knows how Joplin
// stores data: the items table's columns, the JSON envelope inside
// content, and the two-notebook (Session Meetings / Archived) layout. It
// exposes a narrow interface of clean Go values; nothing outside this
// package should import pgx or know a jop_id from a UUID.
package reader

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sfdeloach/my-note-app/agenda-service/parser"
)

const (
	sessionMeetingsTitle = "Session Meetings"
	archivedTitle        = "Archived"
)

// Scope selects which of the two meeting-note notebooks to list.
type Scope int

const (
	Current Scope = iota
	Archived
)

// Note describes one meeting master note as it should appear in a
// listing. Err is non-nil when the note has a per-note problem (bad
// markup_language, a title that doesn't match the naming convention, or a
// title/Metadata disagreement) — the note still appears in listings so
// the problem is visible there, per the brief.
type Note struct {
	ID    string // the note's jop_id
	Title string
	Date  string // parsed from Title; empty if Err is set and parsing failed
	Type  string // parsed from Title; empty if Err is set and parsing failed
	Err   error
}

// NoteBody is a single note's full content, as returned by GetNote.
type NoteBody struct {
	Note
	Body string
}

// Reader answers read-only queries against the Joplin database. Build one
// with New, which also performs the startup-fatal checks described in the
// brief.
type Reader struct {
	pool              *pgxpool.Pool
	sessionMeetingsID string // jop_id
	archivedID        string // jop_id
}

// ErrNoteNotFound is returned by GetNote when the given id doesn't
// resolve to a note inside "Session Meetings" or "Archived" — whether
// because the id is wrong, belongs to something else entirely, or belongs
// to a soft-deleted note. The three cases are deliberately
// indistinguishable to a caller: a guessed jop_id should learn nothing
// about what does or doesn't exist.
var ErrNoteNotFound = errors.New("reader: note not found")

// NoteError describes a per-note problem on an otherwise real, in-scope
// note (as opposed to ErrNoteNotFound, which means there's no such note
// to serve at all).
type NoteError struct {
	ID  string
	Err error
}

func (e *NoteError) Error() string { return fmt.Sprintf("note %s: %v", e.ID, e.Err) }
func (e *NoteError) Unwrap() error { return e.Err }

// New opens against the given pool, runs the startup-fatal checks the
// brief requires (schema shape, notebook resolution, encryption
// invariant), and returns a ready Reader. It never exits the process or
// panics — a non-nil error means the caller (main, in a later stage)
// should refuse to start.
func New(ctx context.Context, pool *pgxpool.Pool) (*Reader, error) {
	if err := checkSchema(ctx, pool); err != nil {
		return nil, fmt.Errorf("reader: startup schema check failed: %w", err)
	}

	var ownerCount int
	if err := pool.QueryRow(ctx, ownerCountQuery).Scan(&ownerCount); err != nil {
		return nil, fmt.Errorf("reader: checking owner_id cardinality: %w", err)
	}
	if ownerCount != 1 {
		return nil, fmt.Errorf("reader: expected exactly one owner_id, found %d", ownerCount)
	}

	sessionMeetingsID, err := resolveNotebook(ctx, pool, sessionMeetingsTitle, notebooksQuery, jopTypeNotebook)
	if err != nil {
		return nil, fmt.Errorf("reader: resolving %q: %w", sessionMeetingsTitle, err)
	}

	archivedID, err := resolveNotebook(ctx, pool, archivedTitle, childNotebooksQuery, jopTypeNotebook, sessionMeetingsID)
	if err != nil {
		return nil, fmt.Errorf("reader: resolving %q: %w", archivedTitle, err)
	}

	if err := checkEncryption(ctx, pool, sessionMeetingsID, archivedID); err != nil {
		return nil, fmt.Errorf("reader: startup encryption check failed: %w", err)
	}

	return &Reader{pool: pool, sessionMeetingsID: sessionMeetingsID, archivedID: archivedID}, nil
}

func checkSchema(ctx context.Context, pool *pgxpool.Pool) error {
	names := make([]string, 0, len(expectedColumns))
	for name := range expectedColumns {
		names = append(names, name)
	}

	rows, err := pool.Query(ctx, schemaCheckQuery, names)
	if err != nil {
		return fmt.Errorf("querying items schema: %w", err)
	}
	defer rows.Close()

	found := make(map[string]string, len(expectedColumns))
	for rows.Next() {
		var name, dataType string
		if err := rows.Scan(&name, &dataType); err != nil {
			return fmt.Errorf("scanning schema row: %w", err)
		}
		found[name] = dataType
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("reading schema rows: %w", err)
	}

	for name, wantType := range expectedColumns {
		gotType, ok := found[name]
		if !ok {
			return fmt.Errorf("items table is missing column %q", name)
		}
		if gotType != wantType {
			return fmt.Errorf("items.%s has type %q, want %q", name, gotType, wantType)
		}
	}
	return nil
}

// resolveNotebook runs query (either notebooksQuery or
// childNotebooksQuery) with the given args, decodes each candidate row's
// content to read its title, and returns the jop_id of the single row
// matching wantTitle. It errors unless exactly one row matches.
func resolveNotebook(ctx context.Context, pool *pgxpool.Pool, wantTitle, query string, args ...any) (string, error) {
	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return "", fmt.Errorf("querying notebooks: %w", err)
	}
	defer rows.Close()

	var matches []string
	for rows.Next() {
		var jopID string
		var content []byte
		if err := rows.Scan(&jopID, &content); err != nil {
			return "", fmt.Errorf("scanning notebook row: %w", err)
		}
		env, err := decodeEnvelope(content)
		if err != nil {
			continue // a malformed notebook row can't be the one we're looking for
		}
		if env.Title == wantTitle {
			matches = append(matches, jopID)
		}
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("reading notebook rows: %w", err)
	}
	if len(matches) != 1 {
		return "", fmt.Errorf("expected exactly one notebook titled %q, found %d", wantTitle, len(matches))
	}
	return matches[0], nil
}

func checkEncryption(ctx context.Context, pool *pgxpool.Pool, notebookIDs ...string) error {
	rows, err := pool.Query(ctx, encryptionCheckQuery, jopTypeNote, notebookIDs)
	if err != nil {
		return fmt.Errorf("querying encryption status: %w", err)
	}
	defer rows.Close()

	var violations []string
	for rows.Next() {
		var jopID string
		if err := rows.Scan(&jopID); err != nil {
			return fmt.Errorf("scanning encryption row: %w", err)
		}
		violations = append(violations, jopID)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("reading encryption rows: %w", err)
	}
	if len(violations) > 0 {
		return fmt.Errorf("%d note(s) under %s/%s have jop_encryption_applied != 0: %v",
			len(violations), sessionMeetingsTitle, archivedTitle, violations)
	}
	return nil
}

// evaluateNote decodes one note row's content and runs every per-note
// check the brief requires: markup_language, title-pattern parsing, and
// the title/Metadata cross-check. skip reports a soft-deleted note, which
// callers must treat as absent, not as an error.
func evaluateNote(jopID string, content []byte) (note Note, body string, skip bool) {
	env, err := decodeEnvelope(content)
	if err != nil {
		return Note{ID: jopID, Err: fmt.Errorf("decoding note: %w", err)}, "", false
	}
	if env.DeletedTime != 0 {
		return Note{}, "", true
	}

	note = Note{ID: jopID, Title: env.Title}

	if env.MarkupLanguage != 1 {
		note.Err = fmt.Errorf("note %q: markup_language %d, want 1 (Markdown)", env.Title, env.MarkupLanguage)
		return note, env.Body, false
	}

	titleDate, titleType, err := parseTitle(env.Title)
	if err != nil {
		note.Err = err
		return note, env.Body, false
	}

	blocks, _, err := parser.Parse(env.Body)
	if err != nil {
		note.Err = fmt.Errorf("note %q: %w", env.Title, err)
		return note, env.Body, false
	}

	metaDate, metaType, err := findMetadata(blocks)
	if err != nil {
		note.Err = fmt.Errorf("note %q: %w", env.Title, err)
		return note, env.Body, false
	}

	if err := crossCheck(titleDate, titleType, metaDate, metaType); err != nil {
		note.Err = fmt.Errorf("note %q: %w", env.Title, err)
		return note, env.Body, false
	}

	note.Date, note.Type = titleDate, titleType
	return note, env.Body, false
}

// ListNotes returns every note in the given scope, sorted by title
// descending. A soft-deleted note is excluded entirely; a note with any
// other problem is still included, with Err set, so the problem is
// visible in the listing rather than causing the note to vanish.
func (r *Reader) ListNotes(ctx context.Context, scope Scope) ([]Note, error) {
	parentID := r.sessionMeetingsID
	if scope == Archived {
		parentID = r.archivedID
	}

	rows, err := r.pool.Query(ctx, listNotesQuery, jopTypeNote, parentID)
	if err != nil {
		return nil, fmt.Errorf("reader: listing notes: %w", err)
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var jopID string
		var content []byte
		if err := rows.Scan(&jopID, &content); err != nil {
			return nil, fmt.Errorf("reader: scanning note row: %w", err)
		}

		note, _, skip := evaluateNote(jopID, content)
		if skip {
			continue
		}
		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reader: reading note rows: %w", err)
	}

	sort.Slice(notes, func(i, j int) bool { return notes[i].Title > notes[j].Title })
	return notes, nil
}

// GetNote fetches a single note by its jop_id. It rejects any id that
// isn't a note inside Session Meetings or Archived with ErrNoteNotFound —
// deliberately without distinguishing a wrong id from one that names
// something real but out of scope, so a guessed jop_id learns nothing.
func (r *Reader) GetNote(ctx context.Context, id string) (NoteBody, error) {
	var content []byte
	err := r.pool.QueryRow(ctx, getNoteQuery, id, jopTypeNote, []string{r.sessionMeetingsID, r.archivedID}).Scan(&content)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return NoteBody{}, ErrNoteNotFound
		}
		return NoteBody{}, fmt.Errorf("reader: fetching note: %w", err)
	}

	note, body, skip := evaluateNote(id, content)
	if skip {
		return NoteBody{}, ErrNoteNotFound
	}
	if note.Err != nil {
		return NoteBody{}, &NoteError{ID: id, Err: note.Err}
	}

	return NoteBody{Note: note, Body: body}, nil
}
