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
- **AI features**: both speech-to-text (Stage 8) and text-to-speech
  (Stage 9) were considered and dropped before implementation. Native
  OS/phone capabilities — dictation, and built-in read-aloud (Android
  Select to Speak, iOS Speak Screen) — already cover the underlying needs
  without adding infrastructure to this project. See Stage 8/Stage 9
  below.
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

## Stage 3 — Backup script + encryption (done)

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

## Stage 4 — Automate + alert (done)

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

## Stage 5 — Restore test (quarterly) (done)

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

## Stage 6 — Lightweight monitoring (done)

**Goal**: catch disk exhaustion, drive health degradation, and container
problems early, without standing up a heavy monitoring stack.

**Tasks**:
- Install `smartmontools`; add a cron check of SMART health attributes
  (reallocated/pending sector counts — this is a spinning HDD, not an SSD,
  so there's no wear-leveling attribute to watch) on the USB-attached HDD.
- Add a cron check of free disk space against a threshold.
- Add a basic container health check (e.g. `docker compose ps` exit
  states).
- Route all of these into the same alert channel used for backups (Stage 4),
  rather than building a separate notification path.

**Verify**: simulate (or wait for) a low-disk-space or degraded-SMART
condition and confirm an alert fires; confirm no alerts in the normal
healthy state.

## Stage 7 — Update strategy (process, not code) (done)

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

## Stage 8 — Speech-to-text for voice notes (dropped)

Scoped originally to transcribe recorded voice notes with whisper.cpp, but
dropped before implementation. Two things drove this: Joplin Server exposes
no Data API (only Desktop/Mobile/CLI clients do), so writing a transcript
back into a note safely would've required running a headless Joplin client
on the Pi just for API access; and on reconsideration, the use case itself
is thin — native OS/phone dictation already covers live note-taking, and
transcribing an already-recorded audio file is rare enough to handle
manually in a separate app rather than building and maintaining a pipeline
for it. See `conversations/12-drop-stage8-speech-to-text.md` for the full
discussion.

## Stage 9 — Text-to-speech for notes (dropped)

Scoped originally as "set up Piper," but the real ask behind it was bigger:
a one-tap play button inside the Joplin app to listen to any note while
driving or doing chores. Joplin's own ecosystem has no read-aloud feature
(an open, unresolved forum request since 2019), and building that true
in-app experience would mean either a custom Joplin plugin plus a small
always-listening HTTP endpoint on the Pi — conflicting with the
no-resident-service and no-reverse-proxy decisions above — or pre-generated
audio attachments, which hits the same headless-Joplin-client-for-API-access
wall Stage 8 did, and isn't spontaneous. Dropped because both Android
("Select to Speak") and iOS ("Speak Screen") already read any on-screen
text aloud, offline, built into the OS, at zero engineering cost — which
covers the actual need. See
`conversations/13-drop-stage9-text-to-speech.md` for the full discussion.

---

**Roadmap resolved**: all 9 stages are now either done or deliberately
dropped. No open stages remain.
