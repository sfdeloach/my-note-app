# 10 — Stage 6 verification (unattended monitoring observation)

**Request:** Confirm Stage 6 ("Lightweight monitoring") is fully working
after the multi-day unattended observation period called for in
`roadmap.md`'s Stage 6 Verify step, and mark the stage done.

**Observed — unattended healthy-state runs:**
- `~/joplin-monitor.log` shows three consecutive nightly runs (Aug 16, 17,
  18, all ~07:45 UTC / 3:45 AM local) in the new log format, each passing
  all three checks: SMART overall health PASSED with all three watched
  attributes (reallocated/pending/uncorrectable sectors) at 0, disk usage
  well under threshold (44-45% vs. 85%), and both containers healthy/
  running. No alerts fired during this window, matching the "confirm no
  alerts in the normal healthy state" half of the Verify step.
- Only three entries present rather than a longer run: the log-format
  change and `logrotate` setup (`09-log-standardization.md`) both landed
  on 8/15, and that session's forced `logrotate -f` test rotation cleared
  the log, so Aug 16 is effectively the first clean entry in the current
  format. No gaps or failures in between.

**Failure-path half of Verify:** already covered during the Stage 6 build
session (`08-stage6-monitoring.md`) rather than repeated here — all three
failure modes (stopped container, lowered disk threshold, loosened SMART
comparison) were manually triggered and confirmed to produce a `/fail`
ping with the correct failure named, then reverted.

**Result:** Both halves of `roadmap.md`'s Stage 6 Verify step are now
confirmed: (1) unattended nightly runs stay clean with no false alerts
across multiple real nights, and (2) all three failure paths (disk, SMART,
container) were confirmed to alert correctly.

**Done:**
- `roadmap.md`: Stage 6 heading changed from "(waiting for 8/18/26 to
  observe behavior)" to "(done)".
