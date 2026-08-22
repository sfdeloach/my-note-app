# Updating Joplin Server & Postgres

This is docs/roadmap.md Stage 7: a manual, deliberate process for keeping the
pinned image tags in `compose.yml` current — never automated, never
`latest`. Run through this checklist periodically (there's no fixed
schedule; treat "it's been a while" or a security advisory as the trigger).

## Checklist

1. **Check what's currently pinned.**
   ```
   grep image: compose.yml
   ```

2. **Check what's newest upstream.**
   - Joplin Server: https://hub.docker.com/r/joplin/server/tags — this is
     the image actually deployed here. Don't use the GitHub releases page
     (github.com/laurent22/joplin/releases) as a signal — it's dominated by
     desktop-app pre-releases and doesn't distinguish the server product.
   - Postgres: https://hub.docker.com/_/postgres/tags?name=\<major\>-trixie
     (e.g. `?name=17-trixie` while on Postgres 17) — stay on `-trixie`
     images to match the Debian base already in use.

3. **Read release notes for anything between the current and target
   version**, not just the target version's own notes — skipped versions'
   changes still apply. Look specifically for: breaking config changes,
   sync-protocol changes (Joplin), and anything flagged as a migration
   step (Postgres).
   - Joplin Server changelog: https://github.com/laurent22/joplin/blob/dev/CHANGELOG.md
   - Postgres release notes: https://www.postgresql.org/docs/release/

4. **Classify the bump:**
   - Joplin Server bump, or a Postgres **minor** bump (e.g. 17.10→17.11):
     low-risk, go to step 5.
   - Postgres **major** bump (e.g. 17→18): the on-disk format isn't
     binary-compatible across majors — go to step 6 instead, skip step 5.

5. **Minor/patch bump (Joplin Server, or Postgres same-major):**
   ```
   # edit compose.yml, bump the tag
   docker compose pull
   docker compose up -d
   docker compose ps          # confirm both services healthy
   ```
   Spot-check the app (log in, open a note, confirm sync still works from
   a client). Also spot-check `agenda-service`: hit `/` over the tunnel/LAN
   and confirm a real note's four views (Agenda, Red-Letter, Minutes,
   Action Items) still render `200` without error — `agenda-service` reads
   the `items` table and its JSON content envelope directly (no Data API
   insulates it), so a Joplin bump changing either could break it silently.
   Check `~/joplin-backup.log` and `~/joplin-monitor.log` after the next
   cron cycle to confirm nothing regressed.

6. **Postgres major bump (dump/reload required):**
   - Run `./scripts/restore-test.sh up` first, as a baseline confirming the
     current backup restores cleanly *before* touching production.
   - Take a fresh manual backup (`./scripts/backup.sh`) so the upgrade has
     a backup from right before the change, not last night's.
   - Edit `compose.yml` to the new Postgres major tag, then follow the
     standard Postgres major-upgrade path: bring up the new major version
     against a **fresh** volume (old major's data files won't mount
     cleanly under a new major), `pg_dump` from the old container,
     restore into the new one, then cut over.
   - Run `./scripts/restore-test.sh up` again afterward against a fresh
     backup taken post-upgrade, confirming the new setup is itself
     restorable — not just that the upgrade succeeded once.
   - `./scripts/restore-test.sh down` to tear down the throwaway stack
     when both runs are done.

7. **Commit the `compose.yml` tag bump** with a message noting what changed
   and why (e.g. citing the release notes item that mattered, if any).

8. **Log the check below**, even if no bump was applied — this is the
   record of when versions were last compared, so "has it been a while?"
   in step 0 has an actual answer.

## Update log

| Date | Joplin Server | Postgres | Action |
|------|---------------|----------|--------|
| 2026-08-18 | 3.7.1 checked — newest available on Docker Hub, no bump | 17.10-trixie checked — 17.11-trixie available (minor, same major) | Dry run only (docs/roadmap.md Stage 7 Verify step) — no changes applied. Postgres minor bump available and low-risk whenever it's convenient to apply. |
