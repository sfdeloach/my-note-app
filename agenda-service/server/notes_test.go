package server

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleList_Current(t *testing.T) {
	s := testServer(t)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, realBadTitleNote) {
		t.Errorf("body missing expected note title %q", realBadTitleNote)
	}
	if !strings.Contains(body, `class="error"`) {
		t.Errorf("body missing an error row for the known bad-title note")
	}
}

func TestHandleList_Archived(t *testing.T) {
	s := testServer(t)

	req := httptest.NewRequest("GET", "/archived", nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestNoteViews(t *testing.T) {
	nvs := noteViews("some id")
	if len(nvs) != len(viewDispatch) {
		t.Fatalf("got %d views, want %d", len(nvs), len(viewDispatch))
	}

	for _, nv := range nvs {
		if !strings.Contains(nv.ViewURL, "id=some+id") {
			t.Errorf("ViewURL %q does not escape the id", nv.ViewURL)
		}
		if nv.Label == "Red-Letter" && nv.DownloadURL != "" {
			t.Errorf("Red-Letter has a DownloadURL: %q, want none", nv.DownloadURL)
		}
		if nv.Label != "Red-Letter" && nv.DownloadURL == "" {
			t.Errorf("%s has no DownloadURL, want one", nv.Label)
		}
	}
}
