# 06 — Stage 4 verification (dead-man's-switch tests)

**Request:** Confirm Stage 4 ("Automate + alert") is fully working after the
multi-day unattended observation period called for in `roadmap.md`'s Stage 4
Verify step, and mark the stage done.

**Observed — unattended rotation:**
- `joplin-backup.log` showed six consecutive nightly runs (08-04 through
  08-09, ~07:15 UTC / 3:15 AM local), each uploading successfully, sending a
  healthcheck ping, and rotating S3 down to exactly 3 objects (FIFO by
  lexicographic timestamp, per the sort logic in `scripts/backup.sh`).
- Two pre-existing manual-test backups from 08-03 (left over from Stage 3
  testing) were correctly rotated out as the first new nightly backups
  arrived, confirming rotation isn't tripped up by an unusual starting state.

**Dead-man's-switch test 1 — passive missed ping (cron silently not
running):**
- Temporarily shortened the healthchecks.io check's schedule (5 min period /
  2 min grace, down from the real 1 day / 1 hour) and commented out the cron
  entry via `crontab -e` so no ping would arrive during the test window.
- First attempt: the check correctly flagged "down," but no notification
  email arrived. Root cause wasn't pinned down precisely, but pointed at the
  email notification channel setup in healthchecks.io rather than the
  check/detection logic itself, since the "down" state was detected
  correctly. Schedule and crontab were reset to normal, the notification
  setup was adjusted, and the test was repeated.
- Second attempt: check flagged "down" and the expected email arrived.
  Schedule and crontab both restored to production values (1 day period / 1
  hour grace; `crontab scripts/backup.cron`).

**Dead-man's-switch test 2 — explicit script failure:**
- Temporarily set `POSTGRES_DATABASE` in `.env` to a nonexistent value and
  ran `./scripts/backup.sh` manually. `pg_dump` failed immediately (first
  command to run after the `EXIT` trap is armed), the trap fired, and the
  `/fail` ping reached healthchecks.io — triggering a down alert email
  without waiting out any grace period.
- `.env` reverted to the real `POSTGRES_DATABASE` value, then a normal
  manual run of `backup.sh` completed successfully, sending a fresh success
  ping and clearing the healthchecks.io alarm.

**Result:** Both halves of `roadmap.md`'s Stage 4 Verify step are confirmed:
(1) unattended rotation holds at exactly 3 snapshots across multiple real
nightly runs, and (2) the dead-man's-switch alert fires correctly for both
failure modes it's meant to catch — cron going silent (passive missed ping)
and the script itself erroring out (explicit `/fail` ping).

**Done:**
- `roadmap.md`: Stage 4 heading changed from "(waiting for 8/7/26 to observe
  behavior)" to "(done)".
