# my-note-app

Self-hosted Joplin note server. See `CLAUDE.md` for architecture decisions
and `roadmap.md` for the stage-by-stage plan and current status — not
repeated here.

## Start / Stop

```
docker compose up -d      # start
docker compose ps         # verify containers healthy
docker compose down       # stop
```

`down` removes the containers but preserves the `joplin_db` named volume, so
routine stop/start is safe. **Never** run `docker compose down -v` for a
routine stop — the `-v` flag deletes the volume, i.e. all notes.

## Restore (catastrophic failure)

`[TO BE REVISITED — no restore procedure exists yet.]` Nothing here is
built: there's no AWS/IAM setup, no backup script, and no backup has ever
been taken (roadmap.md Stages 2, 3). This section will be written once a
backup exists **and** a restore from it has actually been tested end-to-end
(roadmap.md Stage 5) — a scripted-but-untested restore doesn't count.

## Maintenance / Health Checks

`[TO BE REVISITED — no maintenance checks are implemented yet.]`

- Quarterly restore test (roadmap.md Stage 5): the quarterly cadence is
  decided, but the test procedure itself doesn't exist yet (see Restore
  above).
- Ongoing monitoring — SMART/disk-space/container-health cron checks
  (roadmap.md Stage 6): not implemented yet.

This section will list the actual commands/checklist once those stages are
built.
