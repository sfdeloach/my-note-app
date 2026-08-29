package views

import (
	"strings"
	"testing"
)

const validMetadataBody = `# Metadata

## Info

- **date:** 2026-08-11
- **time:** 5:00 PM
- **location:** Classroom 7/8
- **type:** Stated
- **clerk:** RE Kevin Kennedy
- **moderator:** TE Kevin Struyk
`

func TestExtractMetadata_WellFormed(t *testing.T) {
	blocks := parseOrFatal(t, validMetadataBody)

	m, err := extractMetadata(blocks)
	if err != nil {
		t.Fatalf("extractMetadata: unexpected error: %v", err)
	}
	want := meta{
		Date: "2026-08-11", Time: "5:00 PM", Location: "Classroom 7/8", Type: "Stated",
		Clerk: "RE Kevin Kennedy", Moderator: "TE Kevin Struyk",
	}
	if m != want {
		t.Errorf("extractMetadata = %+v, want %+v", m, want)
	}
}

func TestExtractMetadata_NoMetadataSection(t *testing.T) {
	blocks := parseOrFatal(t, "# Notes\n\n## Something\n")
	if _, err := extractMetadata(blocks); err == nil {
		t.Error("extractMetadata: expected error when no Metadata section is present")
	}
}

func TestExtractMetadata_MissingKey(t *testing.T) {
	body := `# Metadata

## Info

- **date:** 2026-08-11
- **time:** 5:00 PM
- **location:** Classroom 7/8
`
	blocks := parseOrFatal(t, body)
	if _, err := extractMetadata(blocks); err == nil {
		t.Error("extractMetadata: expected error when a key is missing")
	}
}

func TestExtractMetadata_MissingClerk(t *testing.T) {
	body := `# Metadata

## Info

- **date:** 2026-08-11
- **time:** 5:00 PM
- **location:** Classroom 7/8
- **type:** Stated
- **moderator:** TE Kevin Struyk
`
	blocks := parseOrFatal(t, body)
	_, err := extractMetadata(blocks)
	if err == nil {
		t.Fatal("extractMetadata: expected error when clerk is missing")
	}
	if !strings.Contains(err.Error(), "clerk") {
		t.Errorf("error = %q, want it to name the missing %q key", err.Error(), "clerk")
	}
}

func TestExtractMetadata_MissingModerator(t *testing.T) {
	body := `# Metadata

## Info

- **date:** 2026-08-11
- **time:** 5:00 PM
- **location:** Classroom 7/8
- **type:** Stated
- **clerk:** RE Kevin Kennedy
`
	blocks := parseOrFatal(t, body)
	_, err := extractMetadata(blocks)
	if err == nil {
		t.Fatal("extractMetadata: expected error when moderator is missing")
	}
	if !strings.Contains(err.Error(), "moderator") {
		t.Errorf("error = %q, want it to name the missing %q key", err.Error(), "moderator")
	}
}
