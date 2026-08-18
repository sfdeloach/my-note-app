# 05 — Stage 4 automate + alert

**Request:** Execute `roadmap.md` Stage 4 ("Automate + alert") — get
`scripts/backup.sh` running unattended nightly via cron, with S3 rotation
and a dead-man's-switch alert so a silent failure (script error *or* cron
itself dying) gets noticed immediately.

**Decisions:**
- gpg passphrase moves from interactive pinentry to a file:
  `/home/steven/.secrets/joplin-backup-passphrase`, mode 600, in a mode-700
  directory outside the repo, read via `gpg --batch --passphrase-file`.
  Deliberately not `--passphrase` (visible via `ps aux`) or an env var
  (readable via `/proc/<pid>/environ`), and deliberately not `.env` — the
  script already copies `.env` into every backup archive, so a passphrase
  stored there would let one leaked plaintext `.env` decrypt every archive.
  Password manager stays the durable off-host copy; the file is a working
  copy for automation only.
- Dead-man's-switch alerting uses both mechanisms healthchecks.io offers:
  a passive missed-ping on a daily schedule (catches "cron stopped firing
  entirely" — nothing runs, so nothing pings, so healthchecks.io notices
  the silence) and an explicit `/fail` ping from an `EXIT` trap (catches
  script failures immediately, without waiting out the grace period). The
  passive path alone satisfies the roadmap requirement; the explicit path
  is a cheap addition for faster diagnosis.
- New env vars `BACKUP_HEALTHCHECK_URL` and `BACKUP_GPG_PASSPHRASE_FILE`
  added to `.env`/`.env.example`, following the same convention as
  `BACKUP_S3_BUCKET`. The ping URL is treated as secret-ish (anyone with it
  can fake a success ping or trigger a nuisance failure alert) but follows
  existing `.env` handling since it's not a credential in the traditional
  sense.
- Rotation logic (list `joplin-backup-*` objects, keep newest 3 by
  lexicographic sort, `aws s3 rm` the rest) runs inside `backup.sh` itself
  after verified upload, not as a separate cron entry — one script stays
  one complete backup cycle. A rotation failure is treated as fatal
  (inherits `set -e`, triggers the `/fail` ping) rather than a silent
  warning, since this project has exactly one alert channel by design and
  Stage 6 monitoring will reuse it — better to over-alert than let bucket
  growth go unnoticed.
- Crontab entry versioned as `scripts/backup.cron` (03:15 local, `flock -n`
  guard against overlapping runs) and installed via
  `crontab scripts/backup.cron` — nothing previously represented the cron
  schedule as a file. Documented caveat: this command replaces the entire
  personal crontab; future one-off changes should go through `crontab -e`
  with the repo file updated to match afterward, not re-run wholesale.
- healthchecks.io "Simple" schedule mode (period + grace) rather than a
  cron expression, to sidestep DST edge cases on a Pi in
  `America/New_York`.

**Done:**
- `scripts/backup.sh`: non-interactive gpg via passphrase file (with a
  fail-fast existence/mode-600 check before touching the database),
  combined cleanup+failure-ping `EXIT` trap, S3 rotation block, success
  ping as the final step.
- `scripts/backup.cron` written (not yet installed into the live crontab).
- `.env` / `.env.example` updated with `BACKUP_HEALTHCHECK_URL` (placeholder
  pending healthchecks.io signup) and `BACKUP_GPG_PASSPHRASE_FILE`.
- `/home/steven/.secrets/` created, mode 700 (passphrase file itself not
  yet created — needs the real passphrase pasted from the password manager).
- `README.md` "Backups" section rewritten to describe automation, rotation,
  passphrase-file requirements, and dead-man's-switch behavior.

**Not yet done (manual steps, outside what I can do unattended):**
- Sign up for healthchecks.io, create the `joplin-backup` check (Simple
  schedule, 1 day period / 1 hour grace), confirm email alerting works, and
  replace the `BACKUP_HEALTHCHECK_URL` placeholder in `.env` with the real
  ping URL.
- Paste the existing password-manager passphrase into
  `/home/steven/.secrets/joplin-backup-passphrase` and `chmod 600` it —
  must reuse the Stage 3 passphrase, not a new one, or existing S3 backups
  become undecryptable.
- Run `./scripts/backup.sh` manually once end-to-end with the real config
  to confirm the non-interactive path, rotation no-op, and success ping all
  work before trusting it to cron.
- Install with `crontab scripts/backup.cron`.
- `roadmap.md` Stage 4 stays unmarked until the above is done and observed
  running unattended for a few days (roadmap's own Verify step needs
  multi-day, real-world observation — not something to mark done from a
  single dry run).

**Verified:** nothing yet — `bash -n` syntax-checked `scripts/backup.sh`
only. Full verification (rotation holding at exactly 3 objects over
several nights; explicit-fail and passive-miss alert paths both firing)
is pending the manual setup steps above.
