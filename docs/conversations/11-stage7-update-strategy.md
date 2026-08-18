# 11 — Stage 7 update strategy

**Request:** Execute `roadmap.md` Stage 7 ("Update strategy") — a written
checklist for deciding when/how to bump the pinned Joplin Server and
Postgres image tags, plus a dry run of checking for available updates.

**Decision:** checklist lives in its own top-level doc, `UPDATING.md`, not
folded into `roadmap.md` or `README.md` — user's explicit preference.

**Checklist contents (`UPDATING.md`):**
- Where to check for newer versions: `hub.docker.com/r/joplin/server/tags`
  for Joplin Server, `hub.docker.com/_/postgres/tags?name=<major>-trixie`
  for Postgres — deliberately *not* the GitHub releases page
  (`github.com/laurent22/joplin/releases`), which is dominated by desktop
  app pre-releases and doesn't distinguish the server product (discovered
  during the dry run below).
- Read release notes for every version between current and target, not
  just the target's own notes.
- Classify the bump: Joplin Server or Postgres minor/patch is low-risk
  (pull, up -d, verify healthy, spot-check); a Postgres **major** bump
  requires the dump/reload path and brackets the upgrade with two
  `restore-test.sh` runs — one beforehand as a baseline, one afterward
  against a fresh post-upgrade backup, per roadmap.md's requirement that
  Stage 5's restore test run before and after any major Postgres upgrade.
- An "Update log" table at the bottom, so future checks have a record of
  when versions were last compared even when no bump was applied.

**Dry run performed (per Stage 7's Verify step):**
- Current pinned versions: `postgres:17.10-trixie`, `joplin/server:3.7.1`
  (from `compose.yml`).
- Checked Docker Hub for both images. Joplin Server: `3.7.1` is the newest
  tag available — no bump possible right now. Postgres: `17.11-trixie` is
  available (current is `17.10-trixie`) — a minor bump, same major
  version, no dump/reload needed.
- No changes applied — this was a dry run only, confirming the checklist
  itself surfaces useful, correct information. The Postgres minor bump is
  logged as available and low-risk whenever it's convenient to apply.

**Done:**
- `UPDATING.md`: new file, full checklist + update log with today's dry
  run entry.
- `README.md`: new "Updates" section pointing to `UPDATING.md`.
- `roadmap.md`: Stage 7 heading changed to "(done)".
