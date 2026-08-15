# 09 — Standardize backup/monitor logs

**Request:** `~/joplin-backup.log` and `~/joplin-monitor.log` (cron output,
outside the repo) needed better information. The backup log was full of
`aws` progress-bar noise and duplicate lines (its own `echo` plus the AWS
CLI's own `upload:`/`delete:` confirmations); the monitor log was too
sparse — a single line, no timestamp, no detail on what was actually
measured.

**Context:** this came up after observing the first two real unattended
runs since Stage 6 went live (2026-08-15, 03:15/03:45 local) — one backup
cycle and the first automatic monitor cycle, both successful per
healthchecks.io. Not a sufficient observation window yet (roadmap target is
8/18/26); this session only touched logging, not Stage 6's status.

**Decisions:**
- Every run — backup or monitor, success or failure — now opens with a
  UTC timestamp header (`Backup for Aug 15, 2026, 7:15:01 UTC` /
  `Health check for Aug 15, 2026, 7:45:01 UTC`) and closes with an
  unambiguous status line (`Backup SUCCESSFUL!` / `FAILURE! One or more of
  the checks did not pass.`) followed by a `------------` separator, so
  entries are easy to scan and to `grep`/split by day.
- Action lines use a `...` prefix (`log()`); sub-detail lines (actual
  values discovered, not just pass/fail) use a `......` prefix (`detail()`
  in `monitor.sh`). This gives `monitor.sh` a real readout — SMART overall
  health plus the raw value of each of the three watched attributes, `/`
  usage against the threshold, and status of both containers — where
  before it only logged a one-line pass/fail with no numbers.
- `aws s3 cp`/`aws s3 rm` in `backup.sh` now pass `--only-show-errors`
  (suppresses progress bars and the redundant `upload:`/`delete:` lines,
  but still surfaces real AWS errors on stderr) — the script's own `log`
  lines (`uploading to ...`, `upload successful`, `rotating out old
  backup: ...`, `deletion successful`) are the single source of truth for
  what happened, instead of duplicating AWS CLI's own narration.
- `backup.sh`'s `EXIT` trap (previously installed partway through, after
  the passphrase/container checks) moved to right after the header line,
  so *every* exit path — including early ones like a missing command or a
  bad `.env` — gets a proper `Backup FAILED!` + `------------` footer
  instead of leaving a bare `ERROR: ...` with no closing line. The `/fail`
  healthcheck ping only fires once `BACKUP_HEALTHCHECK_URL` is actually
  loaded (can't ping a URL that hasn't been read from `.env` yet), so
  very-early failures still get a clean log footer even without a ping.
- `monitor.sh` keeps its existing `FAILURES` array/message text (per your
  note not to undo that work) — failure messages are now also echoed into
  the log itself (`...FAILURE: <message>`) right before the closing line,
  not just posted as the healthcheck `/fail` body, so the local log alone
  is enough to diagnose without needing the healthchecks.io dashboard.
- Neither script's actual check logic, thresholds, cron schedule, or
  `.env` variables changed — this was a logging-format-only pass.

**Done:**
- `scripts/backup.sh`: `log`/`fail` helpers, header/footer via an
  earlier-installed `EXIT` trap, `--only-show-errors` on `aws s3`
  calls, per-step `log` lines replacing raw `echo`s.
- `scripts/monitor.sh`: `log`/`detail` helpers, header/footer, each check
  function now prints what it actually found (not just whether it
  passed), failures echoed to the log before the closing line.

**Verified today:** `bash -n` on both scripts (clean); a standalone dry-run
of the `log`/`detail`/header/footer helpers (not the real checks — no S3
credentials or SMART access from this dev environment) confirmed the
output shape matches the intended format exactly. Full end-to-end
verification (a real backup and a real monitor run producing the new log
format) is pending tonight's 03:15/03:45 cron cycle — not yet observed.

**Not yet done (at the time):** existing log history in
`~/joplin-backup.log` / `~/joplin-monitor.log` was left as-is (old-format
entries); new entries append in the new format starting with the next
cron run.

## Follow-up: log rotation

**Request:** once the new log format was in place, a natural next question
was how to keep `~/joplin-backup.log`/`~/joplin-monitor.log` from growing
unbounded — whether to ever delete them, whether hand-editing was safe, and
what happens if the file is deleted out from under the `>>` cron
redirection.

**Explained (no code changes for this part):**
- Neither script owns its log file — `scripts/backup.cron`'s `>> file
  2>&1` redirection is what routes output to a file at all; run manually,
  the same `echo`/`printf` calls just go to the terminal. Cron has no
  terminal, so without that redirect its default behavior is to email
  output to the crontab owner — not set up here, hence the log files.
- Hand-editing a live log is riskier than it looks: most editors write a
  new file and rename over the original rather than editing in place,
  which detaches the inode a concurrently-running cron job's `>>` file
  descriptor is still writing to. Low-probability given the 03:15/03:45
  schedule, but avoidable by not doing it at all.
- `>>` creates the file if it doesn't exist (identical to `>` in that one
  case) — deleting a log by hand causes zero disruption; the next cron run
  just creates a fresh one. This is the same mechanism that makes
  `logrotate`'s rename-then-let-cron-recreate approach work with no
  `copytruncate`/signal-handling needed, unlike a long-lived daemon that
  holds its log open continuously.

**Decision:** use `logrotate` (already present on Debian, already run
daily via `/etc/cron.daily/logrotate` — no new cron job needed) rather
than a custom trimming script. `weekly` / `rotate 8` / `compress` keeps
~2 months of compressed history. Config lives at
`/etc/logrotate.d/joplin` on the Pi — host config, not versioned in this
repo, same precedent as `/etc/sudoers.d/smartctl-monitor`.

**Done:**
- `README.md`: added a "Log rotation" bullet under Maintenance / Health
  Checks documenting the `/etc/logrotate.d/joplin` config and the
  no-`copytruncate`-needed reasoning.

**Manual steps completed (by user, outside what I can do unattended):**
- Created `/etc/logrotate.d/joplin` with the documented config.
- Verified with `sudo logrotate -d ...` (dry run, no errors).
- Verified with `sudo logrotate -f ...` (forced real rotation) —
  `.1`/`.1.gz` rotated files appeared, original log files gone as
  expected.
- Confirmed a fresh `backup.sh` run recreated `joplin-backup.log` cleanly
  with just the new entry.
- Confirmed `/var/lib/logrotate/status` shows both log paths with today's
  rotation date.

**Verified:** all six setup/verification steps above confirmed working by
the user on the Pi.
