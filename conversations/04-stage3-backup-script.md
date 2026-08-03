# 04 — Stage 3 backup script + encryption

**Request:** Execute `roadmap.md` Stage 3 ("Backup script + encryption") —
a script that produces one complete, encrypted, restorable backup.

**Decisions:**
- New `scripts/` directory at repo root — first script in the repo, sets
  the convention future stages (restore, monitoring) will follow.
- `pg_dump` and a copy of `.env` are bundled into a single tar, encrypted
  once with `gpg --symmetric --cipher-algo AES256`, producing one S3 object
  per run rather than two — avoids Stage 4's future rotation logic having
  to delete a dump/`.env` pair in lockstep.
- Object naming: `joplin-backup-<UTC timestamp>.tar.gpg`
  (`YYYYMMDDTHHMMSSZ`) — UTC avoids Pi clock/DST ambiguity, and the
  lexicographically sortable format lets Stage 4 identify "the 3 newest"
  snapshots by name alone.
- All plaintext (dump, `.env` copy, tar) is written under `mktemp -d`
  (mode 700, outside the repo tree), with an `EXIT` trap set immediately
  after creation so plaintext is wiped on success, on any `set -e` failure,
  or on Ctrl-C.
- `pg_dump` runs via `docker exec` with no `-h` flag, connecting over the
  container's local Unix socket (trusted, no password) — mirrors the
  known-good manual command already documented in `CLAUDE.md`, rather than
  forcing TCP + `PGPASSWORD` plumbing that's never been verified here.
- Container readiness check uses `docker inspect`'s `Health.Status`
  (`compose.yml`'s existing `pg_isready` healthcheck), not just
  "container running" — catches the case where Postgres is still inside
  its `start_period`.
- Passphrase entry is interactive `gpg` pinentry only — no
  `--passphrase`/`--passphrase-fd`/env-var based passphrase. A CLI-arg
  passphrase is visible via `ps aux`; an env-var one is readable via
  `/proc/<pid>/environ` by any process running as the same user. Stage 3 is
  manual-only (cron is Stage 4), so there was no reason to accept either
  weaker option yet.
- New env var `BACKUP_S3_BUCKET` added to `.env`/`.env.example` (value:
  `sfdeloach-joplin-backups`) — not secret, but deployment config, so it
  follows the same `.env` convention as `POSTGRES_*`/`APP_BASE_URL` rather
  than being hardcoded in the script. The gpg passphrase itself deliberately
  stays out of `.env` entirely — it lives only in the password manager,
  since it must never sit next to the secrets (including `.env` itself)
  that it protects.

**Done:**
- `scripts/backup.sh` written and made executable.
- `.env` / `.env.example` updated with `BACKUP_S3_BUCKET`.
- `README.md`: added a "Backups" section; corrected the "Restore" section's
  stale placeholder (previously claimed no AWS/backup work existed at all —
  now reflects Stages 2–3 done, restore still untested pending Stage 5).
- `roadmap.md`: Stage 3 marked `(done)`.

**Verified:**
- `docker compose ps` showed `database` healthy before running.
- `./scripts/backup.sh` ran to completion, prompted for the passphrase
  once, uploaded successfully, and printed the `s3://...` object path.
- `aws s3 ls s3://sfdeloach-joplin-backups/` showed the new object with a
  plausible size.
- Downloaded and `gpg --decrypt`ed the object in a scratch dir outside the
  repo; `tar -tf` listed `dump.sql` and `.env`; `dump.sql` started with the
  expected `-- PostgreSQL database dump` header; `sha256sum` confirmed the
  recovered `.env` matched the live one byte-for-byte (compared via hash,
  not `diff`, to avoid ever printing secret values to the terminal).
- `git status` after the run showed no untracked plaintext in the repo tree
  — confirms the `mktemp -d` + `EXIT` trap approach left nothing behind.
