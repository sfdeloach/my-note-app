# 19 — Quarterly restore test (2026 Q4): agenda_reader role gap found and fixed

**Date**: 2026-11-12

## Context

Quarterly restore test (docs/roadmap.md Stage 5), scheduled for 2026-11-11
and run a day late. First run since 2026-08-11, and the first since the
Agenda Service added the `agenda_reader` Postgres role (session 14).

Pre-flight checks all passed: `compose.restore-test.yml` image tags and
joplin-server env list in sync with `compose.yml` (MAILER_* intentionally
omitted), last nightly backup successful, passphrase file mode 600,
docker/gpg/aws present, production healthy.

## Failure on first run

`./scripts/restore-test.sh up` started the throwaway database but never
started joplin-server. The test database's log showed:

```
ERROR:  role "agenda_reader" does not exist
STATEMENT:  GRANT USAGE ON SCHEMA public TO agenda_reader;
```

Root cause: roles are cluster-wide objects, so `pg_dump` (one database)
never includes them, but it does include the GRANTs that reference them.
`agenda_reader` was created by hand in session 14 and never recorded
anywhere a restore could reach. With `ON_ERROR_STOP=1`, psql aborted at the
first such GRANT; `set -e` then exited before joplin-server started. Data
had fully loaded (26 tables, 371 items) since GRANTs come at the end of the
dump. This would have broken a real disaster recovery in exactly the same
way.

Secondary issue: the script's EXIT trap deletes `psql-restore.log`, so the
failure was silent — output just stopped after "Restoring dump...".

## Fix

- New `scripts/sql/agenda-reader-role.sql`: `CREATE ROLE agenda_reader
  LOGIN PASSWORD :'agenda_password'`. Only the role itself — the dump
  re-applies its grants (`USAGE` on schema public, `SELECT` on `items`),
  matching production's role exactly (verified: plain login role, no other
  attributes or grants).
- `restore-test.sh` runs it before loading the dump, using
  `AGENDA_DB_PASSWORD` from the *backup's* `.env` (keeps the "backup must
  be self-sufficient" rule). Password goes via stdin as a psql `\set`
  (escaped for `\` and `'`; verified round-trip with an awkward test value),
  never as a CLI argument, so it can't appear in `ps`; psql output for that
  step is discarded since an error could echo it.
- On a failed dump restore, the script now prints the last 20 lines of
  `psql-restore.log` before the trap deletes it.
- README restore section: explains that roles aren't in the dump and that a
  real recovery onto a fresh Pi must create `agenda_reader` first.

Rejected alternatives: `pg_dumpall --roles-only` in the backup (would also
include `joplin`, which the Postgres image creates itself, causing a
"role already exists" error on restore); dropping `ON_ERROR_STOP` (would make
this test pass while hiding the next real error).

## Re-run results

- 371 items restored, matching production exactly; 2 user accounts.
- `agenda_reader` has `SELECT` on `items` only; password login over the
  Compose network (scram) succeeded and read 371 items; a write was rejected
  with "permission denied".
- joplin-server responding on 22301 (a `localhost` curl gets the expected
  404 "Invalid origin", since APP_BASE_URL names the Pi's address).
- No decrypted files left in the workdir.
- Time to restorable state: ~90s (vs ~30s in August — larger DB).
- User confirmed in the browser: both accounts present, notes correct. Torn
  down with `restore-test.sh down`.

## Follow-up

- Next restore test due ~2027-02-11.
- If any new Postgres role is ever added, it needs the same treatment as
  `agenda_reader` (versioned SQL + created in the restore path).
