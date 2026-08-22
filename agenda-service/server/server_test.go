package server

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sfdeloach/my-note-app/agenda-service/reader"
)

// testSettingsPath points at the repo's real seeded settings file
// (agenda-service/config/settings.json, Stage 3) — good enough for
// integration tests that only need it to load, not specific roster data.
const testSettingsPath = "../config/settings.json"

// realBadTitleNote is the same real note documented in
// reader/reader_test.go as of 2026-08-21: its title doesn't end in
// "Meeting", so evaluateNote always gives it a non-nil Err.
const realBadTitleNote = "2025-11-08 Called Teleconference"

// testServer builds a Server against a live Postgres instance using
// AGENDA_TEST_DB_* env vars, skipping cleanly if they aren't set — same
// pattern as reader/reader_test.go's testReader.
func testServer(t *testing.T) *Server {
	t.Helper()

	host := os.Getenv("AGENDA_TEST_DB_HOST")
	port := os.Getenv("AGENDA_TEST_DB_PORT")
	name := os.Getenv("AGENDA_TEST_DB_NAME")
	user := os.Getenv("AGENDA_TEST_DB_USER")
	password := os.Getenv("AGENDA_TEST_DB_PASSWORD")
	if host == "" || port == "" || name == "" || user == "" || password == "" {
		t.Skip("AGENDA_TEST_DB_* not set; skipping server integration tests")
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, name)
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		t.Fatalf("creating connection pool: %v", err)
	}
	t.Cleanup(pool.Close)

	r, err := reader.New(context.Background(), pool)
	if err != nil {
		t.Fatalf("reader.New: %v", err)
	}

	return New(r, testSettingsPath)
}
