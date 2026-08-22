package views

import "testing"

func TestFormatDate(t *testing.T) {
	got, err := FormatDate("2026-08-11")
	if err != nil {
		t.Fatalf("FormatDate: unexpected error: %v", err)
	}
	if got != "August 11, 2026" {
		t.Errorf("FormatDate = %q, want %q", got, "August 11, 2026")
	}
}

func TestFormatDate_Malformed(t *testing.T) {
	if _, err := FormatDate("not-a-date"); err == nil {
		t.Error("FormatDate: expected error for a malformed date, got nil")
	}
}
