# 01 — README creation

**Request:** Add a `README.md` as a terse personal operational reference —
start/stop, catastrophic restore (most critical), and routine
maintenance/health checks (folding in the quarterly check mentioned in
`roadmap.md`). No duplication of content already in `roadmap.md` or
`CLAUDE.md` — reference those instead.

**Finding:** The project is earlier than it looked. `compose.yml` still has
the `5432:5432` host port mapping and `.env.example` still has the
`TRANSCRIBE_*` block, so Stage 1 (hardening) hasn't actually been applied
yet, and Stages 2/3/5 (AWS/IAM, backup script, restore test) don't exist at
all. So of the three requested README sections, only Start/Stop has real
content today; Restore and Maintenance are `[TO BE REVISITED]` stubs naming
exactly what's missing and which roadmap stage covers it.

**Decisions:**
- Start/Stop commands are restated directly in the README rather than just
  pointed at CLAUDE.md's Common Commands section — short enough that the
  duplication risk is low, and the README needs to work standalone in an
  emergency. Added a `-v` flag warning on `docker compose down` since that's
  not documented anywhere else and is a real data-loss risk.
- Restore section is a pure stub — no "manual pg_dump as a stopgap" note
  added, even though `CLAUDE.md` already establishes that Database storage
  mode makes a plain `pg_dump` a complete backup. Decided to wait until the
  real backup/restore stages exist rather than half-document an ad-hoc
  interim procedure.

**Done:** `README.md` created at repo root.
