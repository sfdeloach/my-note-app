# 17 — Agenda Service Stage 7: docs & wrap-up (roadmap complete)

**Date**: 2026-08-22

## Context

The user asked to complete Stage 7, the final stage of
`docs/agenda-service/roadmap.md`. Confirmed Stages 1–6 `(done)` in the
roadmap before proceeding (parser `5e7520e`; reader + DB role `7316b22`;
settings `b44d117`; structural views `49c791c`; extractive views `a686f61`;
HTTP layer + Dockerfile + compose `664373f`) — `docker compose ps` on the Pi
also confirmed `agenda-service` already `Up`, publishing `8080:8080`,
alongside `database`/`joplin-server`.

Stage 7's own task list is documentation only — no application code changes
are required by the roadmap. Its one blocking dependency was Stage 6's open
item: no real Joplin note yet complied with the h1/h2/key-value authoring
convention, so the four views had never been smoke-tested against real
content (only the brief's fixture example). During planning, the user
converted a real note — **"2026-08-11 Stated Meeting"**
(id `8681273fdeb04c2aa65e5454fd057db2`) — into the compliant format, which
unblocked the smoke test below.

No conversation entry exists for Stage 5 (extractive views: Minutes + Action
Items, `a686f61`) or Stage 6 (HTTP layer + Dockerfile + compose,
`664373f`) — both were built in sessions that didn't close with a numbered
entry. `docs/agenda-service/roadmap.md`'s own "Decided during implementation"
sections for those stages are the detailed record; this entry doesn't
duplicate them, just notes they happened and links forward from here.

## What was built

- **`agenda-service/README.md`** (new): the authoring convention (title
  pattern, Metadata section, h1/h2/key-value syntax, the valueless-key and
  blockquote rules, which keys each view reads, the `# Absences`
  name-matching rule), how to add a new view (the `Build*Model`/`Render*`
  pattern in `views/`, `viewDispatch` registration in `server/view.go`),
  build/run/test commands, and an env var table.
- **Root `CLAUDE.md`**: opening paragraph now acknowledges `agenda-service/`
  as real Go application code (previously said the repo had none); Project
  Status gained a paragraph on the Agenda Service sub-project and its own
  roadmap; Common Commands gained `agenda-service` build/test/compose
  commands.
- **`UPDATING.md`**: step 5 (minor/patch bump) now includes spot-checking
  `agenda-service` — hit `/` and confirm a real note's four views still
  render `200` — since it reads `items` and the content envelope directly,
  with no Data API to insulate it from a Joplin schema change.
- **`docs/agenda-service/roadmap.md`**: Stage 7 marked `(done)` with a
  confirmed writeup; the "Open items" compliance bullet updated to reflect
  one real note now converted, the rest not blocking.
- **`agenda-service/reader/reader_test.go`**: `TestGetNote` gained a
  `"real note authored in the compliant convention"` subtest against
  `2026-08-11 Stated Meeting`, asserting `GetNote` returns no error and the
  expected `Title`/`Date`/`Type`/non-empty `Body`. This closes a standing
  2026-08-21 TODO comment that had explained why no such fixture existed
  yet — code-adjacent to Stage 7's doc scope, but directly enabled by the
  same note conversion, so it went in alongside the docs rather than as a
  separate follow-up.

## Verification

Ran `go build`, `go vet`, and `go test ./...` for `agenda-service` inside a
`golang:1.25.14-trixie` container attached to the Compose network (Postgres
has no host-published port by design), using the live `agenda_reader`
credentials from `.env`. All packages pass, including the new `TestGetNote`
subtest.

Live smoke test against the running Pi deployment, using the real converted
note (id `8681273fdeb04c2aa65e5454fd057db2`):

- `GET /` lists it with normal view/download links — no inline error row.
- All four views return `200`: Agenda, Red-Letter, Minutes, Action Items.
  Content spot-checked: Agenda's masthead ("Saint Andrew..." /
  "Session Meeting Agenda"); Red-Letter's `Stated Meeting Agenda` head
  matter with an inlined `style="color: firebrick;"` span (no `<style>`
  block — paste-safe); Minutes' Present/Absent rosters and at least one
  Motion; Action Items' at least one "Action Item" entry. No
  error/panic/500 text anywhere.
- Downloads: `session-agenda-2026-08-11.html`,
  `session-minutes-2026-08-11.html`, `session-actionitems-2026-08-11.html` —
  filenames match spec exactly.
  `GET /download?id=...&view=redletter` correctly rejected with `400`.

## Status

Stage 7 marked `(done)` in `docs/agenda-service/roadmap.md`. **All 7 stages
of the Agenda Service roadmap are now complete.** The service is
live-deployed on the Pi, not merely ready for deployment — `docker compose
ps` shows it `Up` alongside `database`/`joplin-server`. Remaining work (not
blocking, tracked in roadmap.md's "Open items"): converting the rest of the
33+ real meeting notes into the compliant authoring convention, as
convenient, so their views render instead of showing a per-note error.
