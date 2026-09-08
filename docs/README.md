# docs/

Historical reference only — not loaded automatically in new sessions (only
`CLAUDE.md` at the repo root is). Open these on demand when past reasoning
is actually needed.

- **`roadmap.md`** — the original 9-stage **infra** build-out plan (Joplin +
  Postgres stack, backups, monitoring, updates). All stages are now done or
  deliberately dropped; kept for the reasoning behind decisions already
  made.
- **`conversations/`** — one numbered file per planning/work session,
  `00-initial-prompt.md` through the final roadmap-closeout entries. The
  durable record of project history.
- **`agenda-service/`** — prompt and design history for the Agenda Service
  **feature** (a Go service rendering four views from a Joplin master note;
  see root `CLAUDE.md`). Organized one numbered `NN-<feature>/` folder per
  feature effort — `01-initial-build/` (the original bring-up) holds
  `prompt.md` (the normative build brief) and `roadmap.md` (its own
  7-stage roadmap, **separate** from the top-level one above — don't
  confuse the two). `appendix/` is a shared reference pool cited across
  features. `README.md` there documents the convention. See root
  `CLAUDE.md` for current state.
