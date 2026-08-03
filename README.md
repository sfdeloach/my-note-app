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

## Backups

`./scripts/backup.sh` produces one complete, encrypted backup: a `pg_dump`
of the database plus a copy of `.env` (needed for a full restore), bundled
into a tar and encrypted with `gpg --symmetric --cipher-algo AES256`, then
uploaded to S3. Requires `.env` present (including `BACKUP_S3_BUCKET`) and
the AWS CLI already configured (roadmap.md Stage 2). You'll be prompted for
the encryption passphrase — kept in a password manager, never in this repo
or the bucket.

```
./scripts/backup.sh
```

Backups land in `s3://sfdeloach-joplin-backups/` as
`joplin-backup-<UTC timestamp>.tar.gpg`. This is manual-only for now — no
cron, rotation, or alerting yet (roadmap.md Stage 4).

## Restore (catastrophic failure)

`[TO BE REVISITED — no restore procedure exists yet.]` AWS/IAM setup
(Stage 2) and the backup script (Stage 3) are complete, and an encrypted
backup exists in S3, but a restore from it has not yet been tested
end-to-end (roadmap.md Stage 5). This section will be written once that
test passes — a scripted-but-untested restore doesn't count.

## Maintenance / Health Checks

`[TO BE REVISITED — no maintenance checks are implemented yet.]`

- Quarterly restore test (roadmap.md Stage 5): the quarterly cadence is
  decided, but the test procedure itself doesn't exist yet (see Restore
  above).
- Ongoing monitoring — SMART/disk-space/container-health cron checks
  (roadmap.md Stage 6): not implemented yet.

This section will list the actual commands/checklist once those stages are
built.
