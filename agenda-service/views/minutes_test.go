package views

import (
	"strings"
	"testing"
)

func TestBuildMinutesModel_ExampleNote(t *testing.T) {
	tree := loadExampleTree(t)
	cfg := loadSeededSettings(t)

	model, err := BuildMinutesModel(tree, cfg)
	if err != nil {
		t.Fatalf("BuildMinutesModel: unexpected error: %v", err)
	}

	if model.Church != churchName {
		t.Errorf("Church = %q, want %q", model.Church, churchName)
	}
	if model.MeetingDate != "August 11, 2026" {
		t.Errorf("MeetingDate = %q, want %q", model.MeetingDate, "August 11, 2026")
	}
	if model.TypeLower != "stated" {
		t.Errorf("TypeLower = %q, want %q", model.TypeLower, "stated")
	}
	if model.Location != "Classroom 7/8" {
		t.Errorf("Location = %q, want %q", model.Location, "Classroom 7/8")
	}
	if model.Time != "5:00 PM" {
		t.Errorf("Time = %q, want %q", model.Time, "5:00 PM")
	}
	if model.Clerk != "RE Kevin Kennedy" {
		t.Errorf("Clerk = %q, want %q", model.Clerk, "RE Kevin Kennedy")
	}
	if model.Moderator != "TE Kevin Struyk" {
		t.Errorf("Moderator = %q, want %q", model.Moderator, "TE Kevin Struyk")
	}

	wantAbsent := "Dave Murray, Burk Parsons"
	if model.Absent != wantAbsent {
		t.Errorf("Absent = %q, want %q", model.Absent, wantAbsent)
	}

	presentNames := strings.Split(model.Present, ", ")
	if len(presentNames) != 16 {
		t.Fatalf("len(Present names) = %d, want 16: %q", len(presentNames), model.Present)
	}
	for _, name := range presentNames {
		if name == "Dave Murray" || name == "Burk Parsons" {
			t.Errorf("Present includes absent elder %q", name)
		}
	}

	wantMotions := []bool{false, true, true, true, false}
	if len(model.Entries) != len(wantMotions) {
		t.Fatalf("len(Entries) = %d, want %d: %+v", len(model.Entries), len(wantMotions), model.Entries)
	}
	for i, want := range wantMotions {
		if model.Entries[i].Motion != want {
			t.Errorf("Entries[%d].Motion = %v, want %v", i, model.Entries[i].Motion, want)
		}
	}

	bankMotion := string(model.Entries[2].Value)
	if !strings.Contains(bankMotion, "<strong>REMOVE</strong>") {
		t.Errorf("Entries[2].Value missing bolded REMOVE: %q", bankMotion)
	}
	if !strings.Contains(bankMotion, "<strong>ADD</strong>") {
		t.Errorf("Entries[2].Value missing bolded ADD: %q", bankMotion)
	}

	// Member Updates: all five tables, always in canonical order.
	wantTitles := []string{"New Members", "Baptisms", "Transfers", "Removals", "Deaths"}
	if len(model.MemberUpdates) != len(wantTitles) {
		t.Fatalf("len(MemberUpdates) = %d, want %d: %+v", len(model.MemberUpdates), len(wantTitles), model.MemberUpdates)
	}
	for i, want := range wantTitles {
		if model.MemberUpdates[i].Title != want {
			t.Errorf("MemberUpdates[%d].Title = %q, want %q", i, model.MemberUpdates[i].Title, want)
		}
	}

	newMembers := model.MemberUpdates[0]
	if len(newMembers.Headers) != 5 || len(newMembers.Rows) != 5 {
		t.Errorf("New Members = %d headers / %d rows, want 5 / 5", len(newMembers.Headers), len(newMembers.Rows))
	}

	deaths := model.MemberUpdates[4]
	wantDeathsHeaders := []string{"First Name", "Middle", "Last Name", "Date"}
	if len(deaths.Headers) != len(wantDeathsHeaders) {
		t.Fatalf("Deaths headers = %d, want %d", len(deaths.Headers), len(wantDeathsHeaders))
	}
	for i, want := range wantDeathsHeaders {
		if string(deaths.Headers[i]) != want {
			t.Errorf("Deaths.Headers[%d] = %q, want %q", i, deaths.Headers[i], want)
		}
	}
	if len(deaths.Rows) != 1 {
		t.Errorf("Deaths rows = %d, want 1", len(deaths.Rows))
	}

	// Cells run through the same Bold pass as motions, so a raw & or > is
	// escaped.
	if got := string(model.MemberUpdates[1].Rows[0][5]); got != "Daniel &amp; Bonnie Peeler" {
		t.Errorf("Baptisms parents cell = %q, want %q", got, "Daniel &amp; Bonnie Peeler")
	}
	if got := string(model.MemberUpdates[3].Rows[0][4]); got != "non-attendance &gt; 1 year" {
		t.Errorf("Removals reason cell = %q, want %q", got, "non-attendance &gt; 1 year")
	}
}

// minutesBodyWithMemberUpdates wraps a member-updates block body (the
// stanzas, without the fences) in an otherwise-minimal compliant note.
func minutesBodyWithMemberUpdates(block string) string {
	return `# Metadata

## Info

- **date:** 2026-08-11
- **time:** 5:00 PM
- **location:** Classroom 7/8
- **type:** Stated
- **clerk:** RE Kevin Kennedy
- **moderator:** TE Kevin Struyk

# Member Updates

` + "```member-updates\n" + block + "\n```\n"
}

func TestBuildMinutesModel_MemberUpdatesUnknownTitleIsHardError(t *testing.T) {
	tree := parseOrFatal(t, minutesBodyWithMemberUpdates("New Membrs\nFirst | Last\nJane | Doe"))
	cfg := loadSeededSettings(t)

	_, err := BuildMinutesModel(tree, cfg)
	if err == nil {
		t.Fatal("BuildMinutesModel: got nil error, want a hard error naming the unknown table title")
	}
	if !strings.Contains(err.Error(), "New Membrs") {
		t.Errorf("error = %q, want it to name the offending title", err.Error())
	}
}

func TestBuildMinutesModel_MemberUpdatesRaggedRowIsHardError(t *testing.T) {
	tree := parseOrFatal(t, minutesBodyWithMemberUpdates("Deaths\nFirst | Last | Date\nJane | Doe"))
	cfg := loadSeededSettings(t)

	_, err := BuildMinutesModel(tree, cfg)
	if err == nil {
		t.Fatal("BuildMinutesModel: got nil error, want a hard error for the row/header cell-count mismatch")
	}
	if !strings.Contains(err.Error(), "Deaths") {
		t.Errorf("error = %q, want it to name the table", err.Error())
	}
}

func TestBuildMinutesModel_TwoMemberUpdatesBlocksIsHardError(t *testing.T) {
	body := minutesBodyWithMemberUpdates("Deaths\nName | Date\nJoe Roe | 2026-01-01") +
		"\n```member-updates\nBaptisms\nName | Date\nAmy Lee | 2026-02-02\n```\n"
	tree := parseOrFatal(t, body)
	cfg := loadSeededSettings(t)

	_, err := BuildMinutesModel(tree, cfg)
	if err == nil {
		t.Fatal("BuildMinutesModel: got nil error, want a hard error for two member-updates blocks")
	}
	if !strings.Contains(err.Error(), "more than one") {
		t.Errorf("error = %q, want it to flag more than one block", err.Error())
	}
}

func TestBuildMinutesModel_MemberUpdatesEmptyTablesRenderNothing(t *testing.T) {
	cfg := loadSeededSettings(t)

	// A table with a header but no data rows is dropped.
	tree := parseOrFatal(t, minutesBodyWithMemberUpdates("Deaths\nName | Date"))
	model, err := BuildMinutesModel(tree, cfg)
	if err != nil {
		t.Fatalf("BuildMinutesModel: unexpected error: %v", err)
	}
	if len(model.MemberUpdates) != 0 {
		t.Errorf("MemberUpdates = %+v, want empty (the only table had no rows)", model.MemberUpdates)
	}

	// A note with no member-updates block at all.
	noBlock := parseOrFatal(t, `# Metadata

## Info

- **date:** 2026-08-11
- **time:** 5:00 PM
- **location:** Classroom 7/8
- **type:** Stated
- **clerk:** RE Kevin Kennedy
- **moderator:** TE Kevin Struyk
`)
	model, err = BuildMinutesModel(noBlock, cfg)
	if err != nil {
		t.Fatalf("BuildMinutesModel: unexpected error: %v", err)
	}
	if model.MemberUpdates != nil {
		t.Errorf("MemberUpdates = %+v, want nil", model.MemberUpdates)
	}
}

func TestBuildMinutesModel_MisspelledAbsenceIsHardError(t *testing.T) {
	const body = `# Metadata

## Info

- **date:** 2026-08-11
- **time:** 5:00 PM
- **location:** Classroom 7/8
- **type:** Stated
- **clerk:** RE Kevin Kennedy
- **moderator:** TE Kevin Struyk

# Absences

## Not A Real Elder
`
	tree := parseOrFatal(t, body)
	cfg := loadSeededSettings(t)

	_, err := BuildMinutesModel(tree, cfg)
	if err == nil {
		t.Fatal("BuildMinutesModel: got nil error, want a hard error naming the unmatched title")
	}
	if !strings.Contains(err.Error(), "Not A Real Elder") {
		t.Errorf("error = %q, want it to name the unmatched title", err.Error())
	}
}

func TestBuildMinutesModel_NoAbsencesSectionMeansNobodyAbsent(t *testing.T) {
	const body = `# Metadata

## Info

- **date:** 2026-08-11
- **time:** 5:00 PM
- **location:** Classroom 7/8
- **type:** Stated
- **clerk:** RE Kevin Kennedy
- **moderator:** TE Kevin Struyk
`
	tree := parseOrFatal(t, body)
	cfg := loadSeededSettings(t)

	model, err := BuildMinutesModel(tree, cfg)
	if err != nil {
		t.Fatalf("BuildMinutesModel: unexpected error: %v", err)
	}
	if model.Absent != "None" {
		t.Errorf("Absent = %q, want %q", model.Absent, "None")
	}
}

func TestRenderMinutes_ExampleNote(t *testing.T) {
	tree := loadExampleTree(t)
	cfg := loadSeededSettings(t)

	var buf strings.Builder
	if err := RenderMinutes(&buf, tree, cfg); err != nil {
		t.Fatalf("RenderMinutes: unexpected error: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"Session Meeting Minutes",
		"Dave Murray, Burk Parsons",
		"<strong>Motion</strong>",
		"<h2>Member Updates</h2>",
		"<h3>New Members</h3>",
		"<th>Received By</th>",
		"<h3>Deaths</h3>",
		"<td>Daniel &amp; Bonnie Peeler</td>",
		"<footer>",
		"Attested by Clerk RE Kevin Kennedy",
		"Attested by Moderator TE Kevin Struyk",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered output missing %q", want)
		}
	}

	// The section sits between the motions and the attestation footer.
	if strings.Index(out, "<h2>Member Updates</h2>") > strings.Index(out, "<footer>") {
		t.Error("Member Updates section renders after <footer>, want before it")
	}
}
