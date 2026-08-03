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
uploaded to S3. Requires `.env` present (including `BACKUP_S3_BUCKET`,
`BACKUP_HEALTHCHECK_URL`, and `BACKUP_GPG_PASSPHRASE_FILE`) and the AWS CLI
already configured (roadmap.md Stage 2).

```
./scripts/backup.sh
```

Backups land in `s3://sfdeloach-joplin-backups/` as
`joplin-backup-<UTC timestamp>.tar.gpg`. Only the 3 most recent snapshots
are kept — the script deletes older ones from S3 after each successful
upload.

**Automation (roadmap.md Stage 4)**: the script runs nightly via cron,
installed from the versioned `scripts/backup.cron`:

```
crontab scripts/backup.cron
crontab -l   # verify
```

This replaces your entire personal crontab — harmless the first time (no
prior crontab exists), but if you ever add another cron job by hand via
`crontab -e`, don't re-run the command above without updating
`scripts/backup.cron` to match first, or you'll silently wipe it.

**Encryption passphrase**: no longer prompted interactively. It's read
non-interactively from the file at `BACKUP_GPG_PASSPHRASE_FILE`
(`/home/steven/.secrets/joplin-backup-passphrase`), which must be mode 600
in a mode-700 directory outside this repo, and must contain the same
passphrase kept in the password manager — never committed to this repo or
stored in the bucket.

**Dead-man's-switch alerting**: each successful run pings the
healthchecks.io check at `BACKUP_HEALTHCHECK_URL`; a failed run pings its
`/fail` endpoint immediately. If cron stops firing at all (nothing runs, so
nothing pings), healthchecks.io itself notices the missed check-in and
alerts by email — this is what catches "cron silently died," not just
"the script errored."

Local run output also lands in `~/joplin-backup.log` for debugging beyond
what the alert email tells you.

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
