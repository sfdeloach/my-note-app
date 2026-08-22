package server

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sfdeloach/my-note-app/agenda-service/reader"
)

func TestHandleView_MissingOrBadParams(t *testing.T) {
	// These all fail before touching the reader, so a nil one is fine.
	s := New(nil, "")

	cases := []struct {
		name string
		url  string
	}{
		{"missing id", "/view?view=agenda"},
		{"missing view", "/view?id=abc"},
		{"unknown view", "/view?id=abc&view=bogus"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", c.url, nil)
			rec := httptest.NewRecorder()
			s.ServeHTTP(rec, req)
			if rec.Code != 400 {
				t.Errorf("status = %d, want 400", rec.Code)
			}
		})
	}
}

func TestHandleDownload_RejectsRedLetter(t *testing.T) {
	s := New(nil, "")

	req := httptest.NewRequest("GET", "/download?id=abc&view=redletter", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestLookupView(t *testing.T) {
	for _, want := range []string{"agenda", "redletter", "minutes", "actionitems"} {
		if _, ok := lookupView(want); !ok {
			t.Errorf("lookupView(%q) not found", want)
		}
	}
	if _, ok := lookupView("bogus"); ok {
		t.Errorf("lookupView(bogus) found, want not found")
	}
}

func TestWriteNoteError(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		rec := httptest.NewRecorder()
		writeNoteError(rec, "abc", reader.ErrNoteNotFound)
		if rec.Code != 404 {
			t.Errorf("status = %d, want 404", rec.Code)
		}
	})

	t.Run("per-note error", func(t *testing.T) {
		rec := httptest.NewRecorder()
		writeNoteError(rec, "abc", &reader.NoteError{ID: "abc", Err: errors.New("title/Metadata mismatch")})
		if rec.Code != 422 {
			t.Errorf("status = %d, want 422", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "title/Metadata mismatch") {
			t.Errorf("body = %q, want it to name the underlying problem", rec.Body.String())
		}
	})

	t.Run("generic error is never shown to the client", func(t *testing.T) {
		rec := httptest.NewRecorder()
		writeNoteError(rec, "abc", errors.New("pq: connection reset by peer, host 10.0.0.5"))
		if rec.Code != 500 {
			t.Errorf("status = %d, want 500", rec.Code)
		}
		if strings.Contains(rec.Body.String(), "10.0.0.5") {
			t.Errorf("body leaked internal error detail: %q", rec.Body.String())
		}
	})
}

// Integration tests below need a live database (AGENDA_TEST_DB_*); they
// skip cleanly without it, same as reader/reader_test.go.

func TestHandleView_NotFound(t *testing.T) {
	s := testServer(t)

	req := httptest.NewRequest("GET", "/view?id=00000000000000000000000000000000&view=agenda", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != 404 {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestHandleView_PerNoteError(t *testing.T) {
	s := testServer(t)

	notes, err := s.rdr.ListNotes(context.Background(), reader.Current)
	if err != nil {
		t.Fatalf("ListNotes: %v", err)
	}
	var badID string
	for _, n := range notes {
		if n.Title == realBadTitleNote {
			badID = n.ID
		}
	}
	if badID == "" {
		t.Fatalf("expected to find %q to test the per-note error path", realBadTitleNote)
	}

	req := httptest.NewRequest("GET", "/view?id="+badID+"&view=agenda", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	if rec.Code != 422 {
		t.Errorf("status = %d, want 422", rec.Code)
	}
}

// TestHandleView_Success (a full 200 render of a compliant note) isn't
// covered yet: as of 2026-08-21 none of the real notes comply with the
// h1/h2/key-value authoring convention (see reader/reader_test.go's
// TestGetNote and the roadmap's "Open items"), so there's no live fixture
// for the happy path. Add it here once a compliant note exists.
