# 14 — Agenda Service Stage 2: `reader` package + DB role (checkpoint 2)

**Date**: 2026-08-21

## Context

Continuing the Agenda Service build (`docs/agenda-service/roadmap.md`).
Stage 1 (parser package, checkpoint 1) turned out to already be built and
merged via PR — just not pulled to this local clone. Pulled it at the
start of this session, then planned and built Stage 2: DB recon, the
`reader` package, and the dedicated read-only DB role.

## What was found (live DB recon)

Queried the running `my-note-app-database-1` container directly and found
two corrections to the brief, now recorded in `roadmap.md`:

- `items` has no `title` or `deleted_time` column. Both live only inside
  the JSON envelope in `content` (bytea) — same as `body` and
  `markup_language`. The row-level `name` column is a synthetic filename
  (`{jop_id}.md`), not the title.
- Soft-deleted rows *are* retained, marked by a nonzero
  `content.deleted_time` in the envelope — confirmed by finding one live.

The brief's two flagged-unverified questions are resolved: soft-delete is
retained (above), and `owner_id` is confirmed single-valued.

Also discovered: **none of the 34 real meeting notes currently comply**
with the h1/h2/key-value authoring convention — they predate it (legacy
freeform headers, markdown tables, an occasional missing colon). This
isn't a bug; it's exactly the scenario the brief's per-note-error design
exists for. Two titles also don't match the `YYYY-MM-DD <Type> Meeting`
pattern at all (`2025-11-08 Called Teleconference`,
`2025-09-16 Called Meeting/DZ Court`), and a template/scratch note also
lives inside the Session Meetings notebook. Flagged in `roadmap.md` since
it affects what Stages 4–6 will show against real data until notes are
authored (or rewritten) in the new convention.

## What was built

`agenda-service/reader/`:
- `envelope.go` — pure JSON-envelope decoding, title-pattern parsing, and
  the Metadata-section lookup/cross-check. No DB access; unit-tested with
  literal data (including a hand-built title/Metadata mismatch case,
  since no real note currently exhibits one — deliberately not tested by
  writing a synthetic row into the live `items` table, which holds real
  session-meeting minutes).
- `queries.go` — SQL strings and the expected-schema map.
- `reader.go` — `Scope`, `Note` (extended with an `Err` field beyond the
  brief's sketch, so a per-note problem surfaces in listings rather than
  causing the note to vanish), `NoteBody`, `Reader`, `New` (runs all the
  startup-fatal checks), `ListNotes`, `GetNote`, and the `ErrNoteNotFound`
  / `NoteError` error types that keep "rejected" vs. "broken but real"
  distinguishable.
- Chose `github.com/jackc/pgx/v5`, used natively rather than through
  `database/sql`, since this service is permanently coupled to
  Postgres/Joplin's schema — the driver-abstraction layer buys nothing.
- `ListNotes` runs the full title/Metadata cross-check on every note (not
  deferred to `GetNote`), matching Stage 3's "always re-read, never
  cache" precedent — the whole point of the check is catching a
  copy-forward drift on the list page, which requires it to be visible
  there.

All tests pass, both standalone (`envelope_test.go`) and against the live
database (`reader_test.go`, integration, env-var-gated, read-only).

## DB role

Presented the `CREATE ROLE agenda_reader` / `GRANT SELECT ON items` SQL
plus a why/tradeoffs write-up (why not reuse `joplin`'s credentials, what
SELECT-only does and doesn't protect against, network implications) —
approved. Created the role, verified it can `SELECT` but a write attempt
is cleanly rejected at the permission layer, and re-ran the full `reader`
test suite against `agenda_reader` to confirm it has exactly the access
the package needs. Password added to `.env` (real value) and
`.env.example` (placeholder, `AGENDA_DB_PASSWORD`) — actual consumption
by `compose.yml` is still Stage 6, per the roadmap's boundary.

## Status

Stage 2 marked `(done)` in `docs/agenda-service/roadmap.md`. Next up:
Stage 3 (settings package).
