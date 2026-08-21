package reader

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Confirmed once, by hand, against the live database on 2026-08-21. These
// only need updating if the two notebooks are ever deleted and recreated.
const (
	wantSessionMeetingsJopID = "4d4ce3e61df64cbe975fe2bd029346cb"
	wantArchivedJopID        = "6917e8df3c414864aaa7686f5e825cfa"

	// A real note whose title doesn't end in "Meeting" at all.
	realBadTitleNote1 = "2025-11-08 Called Teleconference"
	// A real note whose title doesn't end in "Meeting" as its final word.
	realBadTitleNote2 = "2025-09-16 Called Meeting/DZ Court"
)

// testReader builds a Reader against a live Postgres instance using
// AGENDA_TEST_DB_* env vars, skipping cleanly if they aren't set. These
// tests are read-only: they must never write to the database, since
// items holds real session-meeting minutes.
func testReader(t *testing.T) *Reader {
	t.Helper()

	host := os.Getenv("AGENDA_TEST_DB_HOST")
	port := os.Getenv("AGENDA_TEST_DB_PORT")
	name := os.Getenv("AGENDA_TEST_DB_NAME")
	user := os.Getenv("AGENDA_TEST_DB_USER")
	password := os.Getenv("AGENDA_TEST_DB_PASSWORD")
	if host == "" || port == "" || name == "" || user == "" || password == "" {
		t.Skip("AGENDA_TEST_DB_* not set; skipping reader integration tests")
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, name)
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		t.Fatalf("creating connection pool: %v", err)
	}
	t.Cleanup(pool.Close)

	r, err := New(context.Background(), pool)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return r
}

func TestNew_ResolvesNotebooks(t *testing.T) {
	r := testReader(t)

	if r.sessionMeetingsID != wantSessionMeetingsJopID {
		t.Errorf("sessionMeetingsID = %q, want %q", r.sessionMeetingsID, wantSessionMeetingsJopID)
	}
	if r.archivedID != wantArchivedJopID {
		t.Errorf("archivedID = %q, want %q", r.archivedID, wantArchivedJopID)
	}
}

func TestListNotes_Current(t *testing.T) {
	r := testReader(t)

	notes, err := r.ListNotes(context.Background(), Current)
	if err != nil {
		t.Fatalf("ListNotes(Current): %v", err)
	}

	// Documented as of 2026-08-21; the real notebook will keep growing.
	if len(notes) < 34 {
		t.Errorf("got %d notes, want at least 34", len(notes))
	}

	byTitle := make(map[string]Note, len(notes))
	for _, n := range notes {
		byTitle[n.Title] = n
	}

	for _, title := range []string{realBadTitleNote1, realBadTitleNote2} {
		note, ok := byTitle[title]
		if !ok {
			t.Errorf("expected note %q in listing, not found", title)
			continue
		}
		if note.Err == nil {
			t.Errorf("note %q: got nil Err, want a title-pattern error", title)
		}
	}

	for i := 1; i < len(notes); i++ {
		if notes[i-1].Title < notes[i].Title {
			t.Fatalf("notes not sorted descending by title: %q before %q", notes[i-1].Title, notes[i].Title)
		}
	}
}

func TestListNotes_Archived(t *testing.T) {
	r := testReader(t)

	notes, err := r.ListNotes(context.Background(), Archived)
	if err != nil {
		t.Fatalf("ListNotes(Archived): %v", err)
	}
	if len(notes) != 0 {
		t.Errorf("got %d archived notes, want 0 (Archived is currently empty)", len(notes))
	}
}

func TestGetNote(t *testing.T) {
	r := testReader(t)

	// As of 2026-08-21 none of the real notes yet comply with the
	// authoring convention this service introduces (they predate it —
	// legacy headers, tables, missing colons), so there is currently no
	// real note to use as a GetNote success-path fixture. That path is
	// instead covered against the brief's own reference example in
	// envelope_test.go (TestFindMetadata_ExampleNote, TestParseTitle),
	// which is what a compliant note is defined to look like. Once a
	// note is authored in the new convention, a "known-good note" case
	// should be added here.
	notes, err := r.ListNotes(context.Background(), Current)
	if err != nil {
		t.Fatalf("ListNotes(Current): %v", err)
	}

	var badTitleID string
	for _, n := range notes {
		if n.Title == realBadTitleNote1 {
			badTitleID = n.ID
		}
	}
	if badTitleID == "" {
		t.Fatalf("expected to find %q to test the per-note error path", realBadTitleNote1)
	}

	t.Run("notebook id, not a note", func(t *testing.T) {
		_, err := r.GetNote(context.Background(), wantSessionMeetingsJopID)
		if !errors.Is(err, ErrNoteNotFound) {
			t.Errorf("GetNote(notebook id) = %v, want ErrNoteNotFound", err)
		}
	})

	t.Run("nonexistent id", func(t *testing.T) {
		_, err := r.GetNote(context.Background(), "00000000000000000000000000000000")
		if !errors.Is(err, ErrNoteNotFound) {
			t.Errorf("GetNote(nonexistent id) = %v, want ErrNoteNotFound", err)
		}
	})

	t.Run("real note with a non-matching title", func(t *testing.T) {
		_, err := r.GetNote(context.Background(), badTitleID)
		var noteErr *NoteError
		if !errors.As(err, &noteErr) {
			t.Errorf("GetNote(%q) = %v, want a *NoteError", badTitleID, err)
		}
	})
}
