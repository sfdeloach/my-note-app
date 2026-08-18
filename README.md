# my-note-app

Self-hosted Joplin note server. See `CLAUDE.md` for architecture decisions.

**Status: complete, in maintenance mode.** All 9 planned stages are done or
deliberately dropped — see `docs/roadmap.md` and `docs/conversations/` for
the full history, kept for reference only.

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
already configured (docs/roadmap.md Stage 2).

```
./scripts/backup.sh
```

Backups land in `s3://sfdeloach-joplin-backups/` as
`joplin-backup-<UTC timestamp>.tar.gpg`. Only the 3 most recent snapshots
are kept — the script deletes older ones from S3 after each successful
upload.

**Automation (docs/roadmap.md Stage 4)**: the script runs nightly via cron,
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

`./scripts/restore-test.sh up` proves a real S3 backup is restorable: it
downloads the latest (or a given) backup, decrypts it, and restores it into
a throwaway Compose stack — separate project name (`joplin-restore-test`),
separate volume, separate host port (22301) — so it can run safely
side-by-side with production for verification. It deliberately restores
using the `.env` bundled *inside* the backup archive, not the live one, so
this also proves that half of the backup is sufficient for a real disaster
recovery.

```
./scripts/restore-test.sh up [s3-key]   # restore latest backup, or a specific one
```

Wait for "Restore-test stack is up: http://<pi-address>:22301", then log in
and confirm your notes are present and correct. When done:

```
./scripts/restore-test.sh down          # tears down the throwaway stack + volume
```

This is docs/roadmap.md Stage 5, run on a recurring quarterly basis (a calendar
reminder is set). Last verified end-to-end: 2026-08-11 — 243 items restored,
both accounts present, notes confirmed correct in the browser, ~30s to a
restorable state.

**Why a separate compose file**: `compose.restore-test.yml` is a standalone
file, not merged with `compose.yml` via multiple `-f` flags — Compose
concatenates rather than replaces list-type keys like `ports` across merged
files, so an override would make the throwaway stack also try to bind host
port 22300, which production already holds. Keep its image tags in sync with
`compose.yml` by hand if those ever change.

## Maintenance / Health Checks

- Quarterly restore test (docs/roadmap.md Stage 5, done): see Restore above.
- Ongoing monitoring (docs/roadmap.md Stage 6): `./scripts/monitor.sh` checks
  three things — SMART health on the boot/data drive (`/dev/sda`, a
  spinning HDD, not an SSD — see CLAUDE.md), free space on `/` against an
  85% threshold, and container status for both `database` and
  `joplin-server`. It requires passwordless `sudo smartctl` (see
  `/etc/sudoers.d/smartctl-monitor`, scoped to just that binary) since SMART
  data needs root and cron doesn't run as root.

  ```
  ./scripts/monitor.sh
  ```

  Runs nightly via cron at 03:45 (installed the same way as backups, from
  `scripts/backup.cron`), 30 minutes after the backup job so they don't
  compete for disk I/O. Uses the same dead-man's-switch pattern as backups,
  but pings a **separate** healthchecks.io check
  (`MONITOR_HEALTHCHECK_URL` in `.env`) — a monitoring alert and a backup
  alert are distinguishable by which check fired. On failure, the alert
  includes which specific check(s) failed (disk space, SMART, or a named
  container) as the ping body. Local output also lands in
  `~/joplin-monitor.log`.

- **Log rotation**: `~/joplin-backup.log` and `~/joplin-monitor.log` are
  plain cron output (`>> file 2>&1` in `scripts/backup.cron`) — neither
  script manages its own log file. Rotation is handled by `logrotate`
  (already present on Debian, run daily via `/etc/cron.daily/logrotate` —
  no new cron job needed), configured via a drop-in at
  `/etc/logrotate.d/joplin` (host config, not versioned in this repo, same
  as the `/etc/sudoers.d/smartctl-monitor` rule):

  ```
  /home/steven/joplin-backup.log /home/steven/joplin-monitor.log {
      weekly
      rotate 8
      compress
      missingok
      notifempty
  }
  ```

  No `create`/`copytruncate` directive is needed: unlike a daemon that
  holds its log open for its whole lifetime, each cron run opens the file
  fresh and closes it when the script exits, so after `logrotate` renames
  the file out from under it, the next night's cron run just creates a new
  one at that path — the same thing that happens if the file is deleted by
  hand. Keeps roughly 2 months of compressed history (8 weekly rotations)
  before the oldest is dropped.

## Updates

Joplin Server and Postgres image tags in `compose.yml` are bumped manually
and deliberately, never automated, never `latest`. See `UPDATING.md` for
the full checklist (where to check for new versions, what counts as a
low-risk bump vs. a Postgres major version requiring dump/reload, and a
log of past checks).
