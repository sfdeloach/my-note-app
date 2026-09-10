# 18 — Agenda Service: Member Update tables in the Minutes view

**Date**: 2026-09-09

## Context

The user supplied an expert build brief,
`docs/agenda-service/02-member-update-tables/prompt.md`, for a new **Member
Updates** section in the Minutes view: up to five tables (New Members,
Baptisms, Transfers, Removals, Deaths) recording the membership changes
ratified at a meeting, inserted between the motion/comment paragraphs and
the attestation footer. The section renders in the Minutes view only, never
Agenda / Red-Letter / Action Items.

The data is authored as a single fenced ` ```member-updates ` block under a
`# Member Updates` h1 at the end of the master note. The parser previously
treated any code fence — and any Markdown table row — as a hard error that
failed the whole note, so the feature needed one additive parser rule
before the view work.

Work followed the brief's three checkpoints, each shown to the user before
proceeding. Two decisions were settled with the user at the checkpoints:

- **Unterminated fence at EOF** → non-fatal `Warning` + capture-to-EOF,
  matching the valueless-key precedent. A hard error would have undercut
  the side benefit that stray fenced content in a legacy note no longer
  fails the parse.
- **The example note carries all five tables** (columns matching the `09`
  golden doc), so the rendered example is a true byte-for-byte match for
  the golden document rather than an illustrative subset.

## What was built

- **`agenda-service/parser/parser.go`**: one additive rule. A line matching
  `^```(\S*)\s*$` starts verbatim capture; every following line is taken
  as-is (no rule matching, no "unrecognized line" error) until a closing
  `^```\s*$`, then emitted as one top-level (depth-0) `Block` keyed by the
  info string, with the body as its content. Nothing nests under it, so an
  empty `# Member Updates` h1 plus the block are invisible to every
  existing view with no new skip code. An unterminated fence appends a
  `Warning` and captures to EOF. The `member-updates` info string is *not*
  special-cased here — the parser captures any fence; all validation lives
  in the view builder.
- **`agenda-service/parser/parser_test.go`**: new tests for terminated,
  unterminated, info-string (named + empty), otherwise-unrecognized-lines,
  and not-nested-under-h1.
- **`docs/agenda-service/appendix/01-example-note.md`** and
  **`03-scanner-result.md`** (lockstep fixtures): the ad-hoc
  `## Membership Updates` item is replaced with a real `member-updates`
  block carrying all five tables; the expected parse tree is regenerated
  from actual `Parse` output.
- **`agenda-service/views/minutes.go`**: `MemberUpdateTable`
  `{ Title string; Headers []template.HTML; Rows [][]template.HTML }`, a
  `MemberUpdates` field on `MinutesModel`, and `buildMemberUpdates(tree)`.
  It finds the single top-level `member-updates` block, splits stanzas on
  blank lines, validates the title against the fixed set, checks each data
  row's cell count against its header, drops title+header stanzas with no
  rows, orders the tables canonically, and runs every cell through the
  existing `Bold` pass. Hard errors (out of `BuildMinutesModel` → HTTP 422,
  same path as the Absences name check): more than one block, an unknown or
  duplicate title, a missing header row, or a ragged data row.
- **`agenda-service/views/templates/minutes.tmpl`**: a
  `{{if .MemberUpdates}} <h2>Member Updates</h2> {{range …}} <h3>/<table>
  {{end}} {{end}}` block between the `.Entries` loop and `<footer>`, plus a
  table ruleset in `<style>` adapted from the `09` spec. Two CSS-audit
  fixes on top of the user-supplied rules: `h3 { margin-top: 1rem }` (the
  shared `header h1, h2, h3` reset zeroes body headings too) and `td`
  vertical padding `0 → 0.25rem`. A local `@media print { thead {
  display: table-header-group } }` states the page-break header-repeat
  intent.
- **`agenda-service/views/minutes_test.go`**: Member Updates model
  assertions in `TestBuildMinutesModel_ExampleNote` (five tables, canonical
  order, header/row counts, `&`/`>` escaping), render-substring assertions
  plus a before-`<footer>` position check in `TestRenderMinutes_ExampleNote`,
  and hard-error tests for unknown title, ragged row, and two blocks, plus
  an empty-table / no-block test.
- **`docs/agenda-service/appendix/09-member-updates.md`**: added the
  normative preamble (what is / isn't normative, the escaping and
  shared-stylesheet notes) above the golden ` ```html ` document, following
  appendices 05–08; entity-escaped the golden-doc cells that `Bold`
  escapes (`&amp;`, `&gt;`; apostrophe kept literal per the 07 convention).
- **`docs/agenda-service/appendix/07-meeting-minutes.md`**: one-line pointer
  to `09`.
- **`agenda-service/README.md`**: a `# Member Updates` subsection in the
  authoring guide — block format, stanza rules, valid titles, the
  hard-error cases, Minutes-only.
- **`docs/agenda-service/README.md`**: `09-member-updates.md` row in the
  appendix table; `02-member-update-tables/` entry under "What's here now".
- **Root `CLAUDE.md`**: one sentence in the Agenda Service paragraph noting
  the Member Updates section and that the parser now tolerates fenced
  content.

## Verification

`gofmt -l`, `go build ./...`, `go vet ./...`, and `go test ./...` for
`agenda-service` all clean locally (Go 1.24 toolchain; the module declares
1.25.0 but uses no 1.25-only feature). The regenerated
`03-scanner-result.md` round-trips through `TestParseExampleNote`; the
Agenda / Red-Letter / Action Items suites pass unchanged, confirming the
new block is invisible to them with no skip code added. Rendered Minutes
for the example note were dumped and eyeballed: five tables, canonical
order, correct per-table headers, `Daniel &amp; Bonnie Peeler` /
`St. Paul&#39;s PCA` / `non-attendance &gt; 1 year` escaped exactly as
`Bold` produces.

Not yet done (needs the Pi's DB credentials): the live smoke test showing
the section in the deployed Minutes view and nowhere else, and the sweep of
all ~34 real notes confirming none newly hard-errors. The "no new hard
errors" property holds by construction — the capture branch has no error
path and `^```` matches no existing rule — but the sweep is still worth
running.

## Status

Feature code-complete and committed. No standalone `roadmap.md` was written
(a one-to-two-session effort, per `docs/agenda-service/README.md`). Deploy
with `git pull` on the Pi then `docker compose up -d --build agenda-service`,
and run the two deferred checks above.
