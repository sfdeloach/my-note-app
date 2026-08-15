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

**Not yet done:** nothing outstanding from this session. Existing log
history in `~/joplin-backup.log` / `~/joplin-monitor.log` was left as-is
(old-format entries); new entries append in the new format starting with
tonight's run.
