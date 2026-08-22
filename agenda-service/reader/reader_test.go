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

	// The first real note authored in the compliant convention, converted
	// 2026-08-22 (Stage 7). Used as the GetNote success-path fixture.
	realGoodNoteTitle = "2026-08-11 Stated Meeting"
	realGoodNoteDate  = "2026-08-11"
	realGoodNoteType  = "Stated"
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

	notes, err := r.ListNotes(context.Background(), Current)
	if err != nil {
		t.Fatalf("ListNotes(Current): %v", err)
	}

	var badTitleID, goodID string
	for _, n := range notes {
		if n.Title == realBadTitleNote1 {
			badTitleID = n.ID
		}
		if n.Title == realGoodNoteTitle {
			goodID = n.ID
		}
	}
	if badTitleID == "" {
		t.Fatalf("expected to find %q to test the per-note error path", realBadTitleNote1)
	}
	if goodID == "" {
		t.Fatalf("expected to find %q to test the GetNote success path", realGoodNoteTitle)
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

	t.Run("real note authored in the compliant convention", func(t *testing.T) {
		nb, err := r.GetNote(context.Background(), goodID)
		if err != nil {
			t.Fatalf("GetNote(%q) = %v, want no error", goodID, err)
		}
		if nb.Title != realGoodNoteTitle {
			t.Errorf("Title = %q, want %q", nb.Title, realGoodNoteTitle)
		}
		if nb.Date != realGoodNoteDate {
			t.Errorf("Date = %q, want %q", nb.Date, realGoodNoteDate)
		}
		if nb.Type != realGoodNoteType {
			t.Errorf("Type = %q, want %q", nb.Type, realGoodNoteType)
		}
		if nb.Body == "" {
			t.Error("Body is empty, want the note's Markdown body")
		}
	})
}
