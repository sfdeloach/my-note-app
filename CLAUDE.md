# my-note-app

Self-hosted [Joplin](https://joplinapp.org/) note server for personal/work
notes, running as Docker Compose services (Joplin Server + Postgres) on a
Raspberry Pi 5 (4GB RAM, 320GB USB-attached spinning HDD, Debian). Reachable only via a WireGuard
tunnel or trusted home LAN — no public inbound exposure.

There's no application code here to build/lint/test in the traditional
sense; the "work" in this repo is Docker Compose config, `.env` secrets
management, and ongoing maintenance — deliberate image-version bumps
(`UPDATING.md`), backups, monitoring, and the quarterly restore test.

## Project status

Complete, in maintenance mode. All 9 originally-planned stages (hardening →
AWS/IAM setup → backup script → automate/alert → restore test → monitoring
→ update strategy → speech-to-text (dropped) → text-to-speech (dropped))
are done or deliberately dropped. `docs/roadmap.md` has the full
stage-by-stage history; `docs/conversations/` has the numbered session log.
Both are historical reference only — not loaded automatically, read on
demand if past reasoning is needed.

## Architecture decisions already made

These were deliberated in the original planning conversation — don't
re-litigate them without new information:

- **No reverse proxy.** WireGuard already encrypts tunnel traffic; home WiFi
  is fully trusted (no untrusted/guest devices). TLS termination and
  security headers have nothing to protect against here. Revisit only if a
  second service is ever added to the Pi.
- **Storage mode is `Database`.** `STORAGE_DRIVER` is unset, so Joplin puts
  attachments in Postgres, not the filesystem. This means a `pg_dump` alone
  is a complete backup — no separate filesystem sync needed.
- **Both AI feature stages (speech-to-text, text-to-speech) were dropped
  before implementation.** Native OS/phone capabilities — dictation, and
  built-in read-aloud (Android Select to Speak, iOS Speak Screen) — already
  cover the underlying needs without adding a resident service, a reverse
  proxy, or a headless Joplin client (needed for API access since Joplin
  Server itself has no Data API) to this project. See docs/roadmap.md
  Stage 8/Stage 9 for the full reasoning.
- **Monitoring is lightweight and cron-based**, reusing the same alert
  channel as backups — no Prometheus/Grafana.
- **Updates are manual and deliberate, never automated, never `latest`
  tags.** A Postgres major version bump requires a dump/reload since the
  on-disk format isn't binary-compatible across majors.

## Secrets handling

`.env` holds real credentials (Postgres password, Gmail SMTP app password)
and is gitignored. Never print, log, or commit its contents. `.env.example`
is the template — when documenting a new config variable, add it there with
a placeholder, not in `.env`.

## Conversation log

`docs/conversations/` holds one numbered file per planning/work session
(`00-initial-prompt.md` is the first, `13-...` is the most recent — the
roadmap-closeout entries). This is the durable record of project history
and decisions — at the end of a session with meaningful discussion or
decisions, add the next numbered file (`14-...`, etc.) summarizing what was
discussed/decided/done. Expect this to happen less often now that the
project is in maintenance mode.

## Communication style

The user is self-taught and comfortable with Go, Docker, and Postgres, but
weak on networking and security. For anything touching those areas, explain
*why*, not just *what* — give the reasoning and brief tradeoffs rather than
a flat instruction or a single unexplained recommendation.

## Common commands

```
docker compose up -d              # start the stack
docker compose ps                 # check container/health status
docker compose logs -f joplin-server
docker exec -it my-note-app-database-1 pg_dump -U joplin joplin > backup.sql
```
