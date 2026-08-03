# 02 — Stage 1 hardening

**Request:** Execute `roadmap.md` Stage 1 ("Harden the current stack") —
close unnecessary exposure and clean up leftover boilerplate before building
anything new on top.

**Done:**
- `compose.yml`: removed the `5432:5432` host port mapping from the
  `database` service. Postgres is now reachable only over the internal
  `my-note-app-network` bridge network; `joplin-server` still exposes
  `22300:22300` to the host.
- `compose.yml`: added a comment above `services:` recording why there's no
  reverse proxy (WireGuard already encrypts tunnel traffic, home LAN is
  fully trusted, TLS termination/security headers have nothing to protect
  against here) — revisit only if a second service is ever added to the Pi.
- `.env.example`: deleted the commented-out `TRANSCRIBE_*` / `HTR_CLI_*`
  block (handwriting-OCR sidecar boilerplate, unrelated to this project).

**Verified:** `docker compose up -d` started both containers cleanly;
`docker compose ps` confirmed `database` shows no host port binding
(`5432/tcp` internal only) while `joplin-server` still shows
`0.0.0.0:22300->22300/tcp`; Postgres reported healthy; `joplin-server` logs
showed migrations running and a successful DB connection; `/api/ping`
responded (an "Invalid origin" reply, not connection-refused, confirming
the server is reachable on 22300). Actual WireGuard/LAN sync-client
verification from another device is left to the user, per the roadmap's
verify step.

**Note:** at the time this stage was picked up, `compose.yml` already had an
uncommitted, unrelated change in the working tree (explicit `networks:`
blocks added to both services) — left untouched by this work.
