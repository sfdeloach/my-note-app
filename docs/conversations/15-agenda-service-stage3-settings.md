# 15 — Agenda Service Stage 3: `settings` package

**Date**: 2026-08-22

## Context

Continuing the Agenda Service build (`docs/agenda-service/roadmap.md`).
Confirmed Stages 1 and 2 were both `(done)` before starting. This session
ran in a remote/cloud Claude Code environment rather than a local one — it
has no route to the Pi's Postgres instance (WireGuard/LAN only), so before
starting Stage 3 we checked whether that mattered. It didn't: Stage 3 is
pure settings-file parsing and elder-roster logic, with no DB dependency,
so it was safe to build here. (Later stages that do need the live DB —
Stage 6's compose/wiring, most likely — should go through a local instance
instead.)

## What was built

`agenda-service/settings/`:
- `settings.go` — `Settings` (a plain value type, not a pointer — unlike
  `reader.Reader` it holds no live resource), `Elder`, `ListType`,
  `LoadError`, `Load(path) (Settings, error)`, `ListTypeFor`.
- `parse.go` — an unexported `jsonSettings`/`jsonElder` pair mirroring the
  on-disk `{"settings": {...}, "data": {"elders": [...]}}` envelope
  exactly, plus `parse()` and `validate()`. Pure logic, no I/O, mirroring
  how `reader/envelope.go` stays separable from `reader.go`'s DB calls.
- `elders.go` — `Elder.Name()` (no comma, ever), `Elder.IsActiveAt()`
  (inclusive on both `activeAt` and `removedAt`), `ActiveElders()`,
  `SortElders()` (last name, then first) — built now even though nothing
  calls them yet, since Stage 4 (roll-call) and Stage 5 (Present/Absent)
  both need this exact logic.
- Table-driven tests across `parse_test.go`, `settings_test.go`,
  `elders_test.go` — no new dependencies (`encoding/json` + `time` from
  stdlib only, confirmed `go.sum` unchanged).

`agenda-service/config/settings.json` — seeded verbatim from
`docs/agenda-service/appendix/4-settings-and-data.md`'s embedded JSON (20
elders, `elderColumns: 3`).

Decisions made during implementation (recorded in `roadmap.md`'s
"Corrections to the brief" section): dates stay plain ISO strings
throughout (matching `reader.Note.Date`'s existing convention — no
`time.Time` anywhere in the package, `time.Parse` used only once, inside
validation); `Settings` is a value type, not a pointer, matching the
brief's `cfg settings.Settings` by-value view signature; validation adds a
few checks beyond the brief's explicit text (`elderColumns > 0`, non-empty
name/class fields) while deliberately leaving `elderClass`'s value
unrestricted since no view reads it yet.

## Verification

All three roadmap Verify bullets are now automated tests, not manual
checks: `TestLoad_SeededFile` (loads cleanly), `TestLoad_Malformed`
(malformed JSON produces a `*LoadError` naming the file and wrapping the
JSON error, not a crash), `TestActiveElders_SeededSettings` (exactly 18
active elders at `2026-08-11`, excluding Zima and Villavicencio, correctly
sorted). Full module (`parser`, `reader`, `settings`) builds, vets, and
tests clean; `gofmt` clean; `go.sum` untouched.

## Status

Stage 3 marked `(done)` in `docs/agenda-service/roadmap.md`. Next up:
Stage 4 (structural views — Agenda + Red-Letter Agenda).
