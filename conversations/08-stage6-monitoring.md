# 08 — Stage 6 lightweight monitoring

**Request:** Execute `roadmap.md` Stage 6 ("Lightweight monitoring") — cron
checks for disk exhaustion, drive health, and container problems, alerting
through the same kind of dead-man's-switch pattern Stage 4 built for backups.

**Correction found along the way:** `CLAUDE.md` and `roadmap.md` both
described the Pi's storage as a "1TB USB SSD." `smartctl -a /dev/sda` and
`lsblk -f` (run directly on the Pi) showed it's actually a 320GB
USB-attached spinning HDD (WDC WD3200BUCT, 5400rpm) — the only writable
filesystem on the box (`sda2`, mounted at `/`), which `docker info` also
confirmed is Docker's data-root (`/var/lib/docker`). Both docs corrected.
This also changed the SMART check itself: a spinning disk has no
wear-leveling attribute (that's flash-specific), so the check watches
classic HDD failure predictors instead — `Reallocated_Sector_Ct` (ID 5),
`Current_Pending_Sector` (ID 197), `Offline_Uncorrectable` (ID 198).

**Decisions:**
- New, **separate** healthchecks.io check (`MONITOR_HEALTHCHECK_URL`) rather
  than reusing `BACKUP_HEALTHCHECK_URL` — user preference, so a monitoring
  alert and a backup alert are distinguishable by which check fired without
  needing to SSH in first.
- Both healthchecks.io checks (backup and monitor) switched from "Simple"
  schedule (period + grace) to the **Cron** schedule mode, entering the
  exact expressions from `scripts/backup.cron` (`15 3 * * *` / `45 3 * * *`,
  `America/New_York`) — a Simple schedule is a rolling window measured from
  the last ping rather than tied to wall-clock time, so it can't distinguish
  "ran on schedule" from "ran late but within the window." (Stage 4 had used
  Simple mode originally to sidestep DST edge cases; revisited here since
  Cron mode handles DST fine and is more precise.)
- `scripts/monitor.sh` deliberately does **not** use `set -e` for the check
  section, unlike `backup.sh` — it needs to run all three checks
  (SMART/disk/containers) regardless of whether an earlier one fails, so one
  bad reading doesn't mask the other two. Failures accumulate in an array;
  a single ping (success or `/fail` with the failure list as the body) goes
  out at the end.
- `smartctl` needs root to read SMART data, but cron runs as the user, not
  root — solved with a sudoers rule scoped to just that one binary
  (`steven ALL=(root) NOPASSWD: /usr/sbin/smartctl` in
  `/etc/sudoers.d/smartctl-monitor`) rather than a blanket `NOPASSWD: ALL`,
  so a compromised cron job can't escalate beyond running `smartctl`.
- Disk-space threshold (85%) and the SMART device path/attribute IDs are
  hardcoded constants near the top of `monitor.sh`, not `.env` vars —
  following the existing convention (`KEEP=3` in `backup.sh`) that
  non-secret, Pi-specific tuning values live in the script, not the
  environment file.
- Cron entry offset 30 minutes after the backup job (03:45 vs 03:15) so they
  don't compete for disk I/O; own `flock` lock file and log, same pattern as
  the backup line.

**Done:**
- `scripts/monitor.sh`: SMART health check (`smartctl -H` + three attribute
  raw values), disk-space check (`df` against `/`), container-health check
  for both `database` and `joplin-server` (reusing the `docker inspect`
  pattern from `backup.sh`), single success/`fail` healthcheck ping at the
  end.
- `.env.example`: added `MONITOR_HEALTHCHECK_URL`.
- `scripts/backup.cron`: added the 03:45 monitoring line.
- `README.md`: "Maintenance / Health Checks" section filled in (was a
  `[TO BE REVISITED]` stub).
- `CLAUDE.md` / `roadmap.md`: storage description corrected from "1TB USB
  SSD" to "320GB USB-attached spinning HDD"; Stage 6 task wording changed
  from "SMART wear-leveling" to "SMART health attributes."

**Manual steps completed (by user, outside what I can do unattended):**
- Created the `joplin-monitoring` healthchecks.io check, set to Cron
  schedule matching `crontab -l`, and put its ping URL in the real `.env`.
- Added the `/etc/sudoers.d/smartctl-monitor` rule.
- Reinstalled the crontab (`crontab scripts/backup.cron`).
- Ran `./scripts/monitor.sh` manually in the healthy state — all three
  checks passed, success ping confirmed received on healthchecks.io.
- Manually tested all three failure paths:
  1. Stopped the `database` container, ran `monitor.sh`, confirmed the
     `/fail` ping fired naming the container, restarted it.
  2. Temporarily lowered `DISK_SPACE_THRESHOLD_PCT`, confirmed the `/fail`
     ping fired with a disk-usage message, restored the real threshold.
  3. Temporarily loosened the SMART raw-value comparison, confirmed the
     `/fail` ping fired naming a SMART attribute, restored it.

**Not yet done:** the roadmap's own Stage 6 Verify step calls for observing
this running unattended for a few days (same reasoning as Stage 4) — that
period is starting now, not yet elapsed.

**Verified today:** `bash -n` syntax check on `monitor.sh` (`shellcheck` not
installed on the Pi, so no lint pass); all three manual failure-path tests
above, confirmed both in script output and on the healthchecks.io dashboard.
Multi-day unattended cron observation is pending — a follow-up conversation
entry will close this out once that's done, mirroring how Stage 4 was
closed (`06-stage4-verification.md`).

**Result:** `roadmap.md` Stage 6 heading changed from "Lightweight
monitoring" to "Lightweight monitoring (waiting for 8/18/26 to observe
behavior)" — not yet marked `(done)`.
