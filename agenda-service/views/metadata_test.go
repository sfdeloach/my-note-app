package views

import "testing"

const validMetadataBody = `# Metadata

## Info

- **date:** 2026-08-11
- **time:** 5:00 PM
- **location:** Classroom 7/8
- **type:** Stated
`

func TestExtractMetadata_WellFormed(t *testing.T) {
	blocks := parseOrFatal(t, validMetadataBody)

	m, err := extractMetadata(blocks)
	if err != nil {
		t.Fatalf("extractMetadata: unexpected error: %v", err)
	}
	want := meta{Date: "2026-08-11", Time: "5:00 PM", Location: "Classroom 7/8", Type: "Stated"}
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
