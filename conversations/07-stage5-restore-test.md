# 07 — Stage 5: restore test

**Request:** Build and run the quarterly restore test called for in
`roadmap.md`'s Stage 5 — prove a real S3 backup is actually restorable, on a
repeatable, scripted, non-destructive basis.

**Built:**
- `compose.restore-test.yml` — a standalone throwaway stack (database +
  joplin-server), not merged with production's `compose.yml` via multiple
  `-f` flags, because Compose concatenates rather than replaces list-type
  keys like `ports` across merged files — an override would have made the
  throwaway stack also try to bind host port 22300, which production already
  holds. Runs under its own Compose project name (`joplin-restore-test`, set
  by the script), so its `joplin_db` volume resolves to
  `joplin-restore-test_joplin_db` — fully separate from production's
  `my-note-app_joplin_db`. Also sets `restart: "no"` (never survives a
  reboot or lingers if forgotten) and forces `MAILER_ENABLED=0` (so poking
  around the restored instance can never send real email through
  production's SMTP creds).
- `scripts/restore-test.sh up [s3-key]` / `down` — downloads the latest (or a
  specified) S3 backup, decrypts it, brings up only `database` first, waits
  for healthy, restores the dump via `psql --set ON_ERROR_STOP=1`, then
  brings up `joplin-server` (in that order, so Joplin never auto-creates its
  own schema against an empty DB before the real dump lands). Restores using
  the `.env` bundled *inside* the backup archive, not the live one — a real
  disaster recovery only has what's in the backup, so this is also what
  proves that half of the backup is sufficient. An `EXIT` trap deletes the
  decrypted `dump.sql`/`.env` on every exit path so plaintext secrets never
  outlive the run. Prints elapsed time to a restorable state as an informal
  RTO estimate.

**Two real bugs found and fixed by actually running it against production's
latest backup:**
1. Joplin Server validates each request's origin against `APP_BASE_URL` and
   404s on a mismatch. The restored `.env` correctly carries production's
   `APP_BASE_URL` (port 22300), but the throwaway stack is deliberately
   reachable on a different host port (22301) to coexist with production on
   the same Pi. Fixed by rewriting just the port in `APP_BASE_URL` (not
   `APP_PORT`, which must stay 22300 to match the container-internal side of
   the `22301:22300` mapping) in the extracted `.env` before starting
   `joplin-server`.
2. The above fix initially had no effect, for a more interesting reason: the
   script loaded `BACKUP_S3_BUCKET`/`BACKUP_GPG_PASSPHRASE_FILE` from
   production's `.env` via `set -a; source ...; set +a`, which exports
   *every* variable in that file into the shell — and Compose gives
   shell-exported variables priority over `--env-file` during
   interpolation. So production's `APP_BASE_URL` (and every other var) was
   silently overriding the restored backup's values. Harmless for a same-day
   backup (values happened to match), but for an older backup — e.g. one
   from before a password rotation — this would have meant the test silently
   validated *today's* credentials instead of what the backup actually
   contains, defeating the point of the test. Fixed by extracting only the
   two values the script needs from production's `.env`, via `grep`/`cut`,
   instead of blanket-sourcing the whole file.

**Verified (real run, not a dry run):** restored `joplin-backup-20260811T071501Z.tar.gpg`
end to end. `items` table showed 243 rows; both real user accounts
(`admin@localhost`, `sfdeloach@gmail.com`) present; `/api/ping` returned
`{"status":"ok"}`; logged in via browser and confirmed real notes were
present and correct. Time to restorable state: ~30s. Throwaway stack torn
down afterward (`docker compose ... down -v`, scoped to the separate project
name); production confirmed untouched and still running throughout.

**Done:**
- `roadmap.md`: Stage 5 heading changed to "(done)".
- `README.md`: Restore section rewritten with the real procedure, replacing
  the `[TO BE REVISITED]` placeholder.
- A recurring quarterly calendar reminder was created to re-run this test.
