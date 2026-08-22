# 16 — Agenda Service Stage 4: structural views (Agenda + Red-Letter Agenda)

**Date**: 2026-08-22

## Context

The user asked to "complete Stage 3," but `docs/agenda-service/roadmap.md`
already showed Stages 1–3 as `(done)` (parser `5e7520e`, reader + DB role
`7316b22`, settings `b44d117`). Verified this against the actual code (build
clean, tests pass, `git log` confirms) rather than trusting the roadmap text
alone, then asked the user how to proceed; they confirmed the real next step
was Stage 4. Like Stage 3, Stage 4 has no DB dependency — both views are
pure functions of a parsed note tree plus `settings.Settings` — so it was
built without a live Postgres connection.

## What was built

New `agenda-service/views/` package:
- `bold.go` — `Bold(s string) template.HTML`: HTML-escapes first via
  `template.HTMLEscapeString`, then substitutes paired `**…**` into
  `<strong>` with a non-greedy regexp; an unmatched `**` is left literal.
  Shared building block, reused by Minutes' bank motion in Stage 5.
- `dateformat.go` — `FormatDate(iso string) (string, error)`: parses
  `YYYY-MM-DD` as UTC, formats `January 2, 2006`. Shared, reused by Minutes.
- `metadata.go` — unexported `extractMetadata`, reads the `# Metadata`
  section's `date`/`time`/`location`/`type`. Views need this independently
  of `reader` since the brief's view signature takes only the tree and
  settings, not the already-validated `reader.Note`.
- `agenda.go` — `AgendaModel`, `BuildAgendaModel`, `RenderAgenda`: masthead,
  roll-call band (active elders, sorted, `elderColumns` CSS columns,
  CSS-drawn checkboxes), body sections as item-title-only `<ol>`/`<ul>` per
  `listType`. No key-values render.
- `redletter.go` — `RedLetterModel`, `BuildRedLetterModel`,
  `RenderRedLetter`: independent head matter (no church name, no roll
  call), each item's `redLetter` value(s) wrapped `(Note: …)` in a span with
  the red color inlined on `style` (survives email clients stripping
  `<style>` blocks on paste), a `#copy-content` container plus one small
  `<script>` for the Clipboard-API copy button — the brief's only permitted
  client-side JS.
- `render.go` — `embed.FS` + `template.ParseFS` over `templates/*.tmpl`.
- `templates/print.tmpl` — the shared print stylesheet (`@page`, screen-desk
  chrome, `@media print` reset), parameterized by `.FontStack`
  (`template.CSS`, not `string` — the contextual CSS escaper otherwise
  mangles a plain string's quotes/commas) so Stage 5's Minutes/Action Items
  can reuse it with the Cambria stack without editing this file.
- `templates/agenda.tmpl`, `templates/redletter.tmpl`.
- Table-driven tests across 5 `_test.go` files (14 test functions, all
  passing) — including `TestBuildAgendaModel_EmptySectionOmitted` and its
  Red-Letter counterpart (a hand-built fixture, since the real example note
  has no empty section), and
  `TestRenderRedLetter_CopyContentSurvivesWithoutStyleBlock`, which
  automates the paste-survival check by asserting the rendered
  `#copy-content` subtree has inline `style="color: firebrick;"` and no
  `<style>` tag.

Decisions made during implementation (recorded in `roadmap.md`'s "Decided
during implementation in Stage 4" list): Metadata extraction lives in
`views`, not `reader`, since views don't receive `reader.Note`; `churchName`
stays a private const in `agenda.go` until Stage 5's Action Items needs it
too; Agenda's and Red-Letter's section/item walks are two small separate
functions rather than one forced-generic walk; bold substitution uses a
non-greedy regexp rather than manual delimiter scanning.

## Verification

`go build ./...`, `go vet ./...`, `gofmt -l .`, `go test ./...` all clean
across the whole module. Both views rendered against the example note +
seeded `config/settings.json` and inspected as raw HTML: Agenda's DOM
structure and content match appendix 5 (masthead, 18-elder roll call in 3
columns excluding Zima/Villavicencio, five body sections in order, Absences
correctly `<ol>`); Red-Letter's head matter and per-item `(Note: …)` spans
match appendix 6 exactly, including bare rendering for "Burk Parsons" and
"Bank Authorization" (no `redLetter` key). Sent the two rendered files to
the user for a pixel-level visual check in an actual browser — that and a
real email-client paste test remain manual follow-ups, since neither is
automatable from this environment.

## Status

Stage 4 marked `(done)` in `docs/agenda-service/roadmap.md`. Next up: Stage
5 (extractive views — Minutes + Action Items), which reuses the bold-inline
renderer, print stylesheet, and Metadata extraction built here.
