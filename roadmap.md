# Roadmap: Self-Hosted Joplin Server

This roadmap breaks the project into stages you can execute one at a time in
future step-by-step conversations. Each stage has a **Goal**, a rough list of
**Tasks**, and a **Verify** step to confirm it worked before moving on.

Stages are ordered by priority: hardening the existing stack first, then
disaster recovery (backups), then ongoing maintenance (monitoring/updates),
then the optional AI features layered on top last — protecting the data that
already exists matters more than adding new capability.

## Decisions already made (context for all stages)

- **Reverse proxy**: skipping nginx. WireGuard already encrypts tunnel
  traffic; home WiFi is fully trusted (no untrusted/guest devices); security
  headers and future-proofing don't justify the added surface area for a
  single-service, no-public-inbound setup. Revisit only if a second service
  ever gets added to the Pi.
- **Storage mode**: `STORAGE_DRIVER` is unset, so Joplin defaults to
  `Database` mode — all notes and attachments live in Postgres. A `pg_dump`
  is a complete backup; no separate filesystem attachment sync is needed.
- **AI features**: both speech-to-text and text-to-speech are real, wanted
  use cases (not speculative), scoped as on-demand batch jobs — not resident
  services — given the 4GB RAM budget already shared by Joplin + Postgres.
- **Monitoring**: lightweight cron-based checks only, reusing the backup
  alert channel — no Prometheus/Grafana, too heavy for this hardware.
- **Updates**: fully manual and deliberate, not automated.

---

## Stage 0 — Baseline (done)

Already complete: `compose.yml` with Joplin Server 3.7.1 + Postgres
17.10-trixie, versions pinned (not `latest`), stack starts cleanly, WireGuard
tunnel configured, admin/user account creation verified by email.

## Stage 1 — Harden the current stack (done)

**Goal**: close unnecessary exposure and clean up leftover boilerplate before
building anything new on top.

**Tasks**:
- Remove the `5432:5432` host port mapping from `compose.yml` for the
  `database` service — Joplin reaches Postgres over the internal Docker
  network, so no host port is needed.
- Remove the unused, commented-out `TRANSCRIBE_*` / `HTR_CLI_*` block from
  `.env.example` (handwriting-OCR sidecar boilerplate, unrelated to this
  project).
- Add a short comment near the top of `compose.yml` (or a note in this repo)
  recording *why* there's no reverse proxy, so it doesn't get re-litigated
  later.

**Verify**: `docker compose up -d` starts cleanly; `docker compose ps` shows
no host binding for port 5432; Joplin is still reachable on 22300 via
WireGuard and LAN; an existing Joplin sync client can still sync.

## Stage 2 — AWS CLI + IAM setup (done)

**Goal**: get the Pi able to talk to S3 with credentials scoped to only what
backups need.

**Tasks**:
- Install AWS CLI v2 on the Pi.
- Create (or designate) an S3 bucket dedicated to these backups.
- Create a narrowly-scoped IAM policy allowing only `s3:PutObject`,
  `s3:GetObject`, and `s3:ListBucket` on that one bucket, and an IAM user/role
  using it — not broad account credentials.
- Configure the AWS CLI on the Pi with those scoped credentials.

**Verify**: `aws s3 ls s3://<bucket>` works with the new credentials; a
manual test file can be uploaded and then deleted from the bucket.

## Stage 3 — Backup script + encryption

**Goal**: a script that produces one complete, encrypted, restorable backup.

**Tasks**:
- Write a script that runs `pg_dump` via `docker exec` into the Postgres
  container (guarantees version match with the server).
- Encrypt the dump and a copy of `.env` (needed for a full restore) with
  `gpg --symmetric --cipher-algo AES256`, using a passphrase kept in a
  password manager — never committed to the repo or stored in the bucket.
- Upload the encrypted files to S3; clean up local plaintext temp files
  immediately after.

**Verify**: run the script manually once; confirm the encrypted object lands
in S3; confirm it can be decrypted locally with the passphrase
(`gpg --decrypt`).

## Stage 4 — Automate + alert

**Goal**: backups run unattended, and you find out immediately if they stop
working — including if cron itself silently stops firing.

**Tasks**:
- Wire the Stage 3 script into cron, running nightly.
- Implement rotation: keep the last 3 daily snapshots in S3, delete older
  ones.
- Add a dead-man's-switch style ping (e.g. a healthchecks.io-style monitor)
  that the script pings on success — you get alerted if the expected ping
  doesn't arrive, catching both script failures and cron not running at all.

**Verify**: let cron run unattended for a few days and confirm S3 shows
exactly 3 rotating snapshots; deliberately disable the cron job once and
confirm the dead-man's-switch alert fires.

## Stage 5 — Restore test (quarterly)

**Goal**: prove the backup is actually restorable, on a repeatable schedule.

**Tasks**:
- Write a scripted process: spin up a throwaway Docker Compose project
  (different project name / volume than production), decrypt the latest S3
  backup, restore the dump into a fresh Postgres container, point a Joplin
  instance at it.
- Confirm real notes are visible and correct, then tear the throwaway stack
  down.
- Put a recurring quarterly reminder in your own calendar/task system to
  re-run this.

**Verify**: notes appear correctly in the ephemeral instance; note roughly
how long the whole process took, as an informal recovery-time estimate.

## Stage 6 — Lightweight monitoring

**Goal**: catch disk exhaustion, SSD wear, and container problems early,
without standing up a heavy monitoring stack.

**Tasks**:
- Install `smartmontools`; add a cron check of SMART wear-leveling
  attributes on the USB SSD.
- Add a cron check of free disk space against a threshold.
- Add a basic container health check (e.g. `docker compose ps` exit
  states).
- Route all of these into the same alert channel used for backups (Stage 4),
  rather than building a separate notification path.

**Verify**: simulate (or wait for) a low-disk-space or degraded-SMART
condition and confirm an alert fires; confirm no alerts in the normal
healthy state.

## Stage 7 — Update strategy (process, not code)

**Goal**: keep Joplin and Postgres current without risking data loss from an
unplanned major-version jump.

**Tasks**:
- Write a short checklist: periodically check Joplin Server and Postgres
  release notes, decide whether to bump the pinned image tag in
  `compose.yml`, and apply it deliberately (never `latest`).
- Explicitly call out that a Postgres **major** version bump (e.g. 17→18)
  requires a dump/reload — the on-disk format isn't binary-compatible across
  majors, unlike minor version bumps within Postgres 17.
- Require running Stage 5's restore test before and after any major Postgres
  upgrade.

**Verify**: checklist exists as a document; do one dry run of checking for
available updates using it.

## Stage 8 — Speech-to-text for voice notes

**Goal**: get voice notes transcribed to searchable text, without a resident
service eating into the 4GB RAM budget.

**Tasks**:
- Check the current Joplin plugin ecosystem for existing Whisper-based
  transcription support before building anything custom.
- If nothing sufficient exists, set up whisper.cpp (tiny or base model) as an
  on-demand script/job — triggered manually or on a schedule, not always
  running.

**Verify**: transcribe a real voice note; check transcription accuracy and
speed are acceptable, and confirm memory usage stays within budget for the
duration of the job (with Joplin + Postgres also running).

## Stage 9 — Text-to-speech for notes

**Goal**: have notes read back to you, again without a resident service.

**Tasks**:
- Check the Joplin plugin ecosystem for existing TTS support first.
- If needed, set up Piper as an on-demand script that generates audio for a
  given note.

**Verify**: generate and play back audio for a sample note; confirm quality
and resource usage are acceptable as an on-demand job.
