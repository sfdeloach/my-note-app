# my-note-app

This is the initial planning conversation for a new project. Interview me to fill in missing information, and give recommendations with brief tradeoffs (not just a single answer) wherever a decision has more than one reasonable path. I'm a self-taught programmer comfortable with Go, Docker, and Postgres, but I'm weak on networking and security — so for anything touching those areas, explain *why*, not just *what*.

Ask questions in small batches (2-3 at a time), not all at once, and don't move to the next topic until the current one is resolved.

**Output:** Once the interview is complete, produce `roadmap.md` — a planning document that lays out the project in stages I can execute one at a time in future step-by-step conversations. Each stage should have a clear goal, a rough list of tasks, and how to verify it worked before moving on. Don't include a full `CLAUDE.md` yet — that comes later, once the roadmap is settled.

## Project: Self-hosted Joplin Server

**Environment (already in place):**
- Raspberry Pi 5, 4 GB RAM, 1 TB USB SSD, Debian, Docker + Docker Compose installed
- WireGuard already configured; the Pi is reachable only via WireGuard tunnel or local home WiFi — no other inbound exposure
- `compose.yml` already written and tested: Joplin Server (official image) + Postgres, stack starts cleanly, I can log into the Users/admin page and create + verify accounts by email

**Topics to cover in the interview:**

1. **Reverse proxy decision.** Given the network is WireGuard/LAN-only with no public inbound exposure, evaluate whether an Nginx reverse proxy is actually adding anything here (e.g., TLS termination for LAN traffic, HTTP security headers, future-proofing if I ever expose a second service on the Pi) versus just being extra surface area with nothing to protect against. Don't just confirm my assumption — walk through the actual reasoning and give a recommendation.

2. **AI/transformer features — scoped to what a 4GB Pi 5 can realistically do.** Treat a general local LLM (chat/summarization) as out of scope for this hardware alongside a running Joplin+Postgres stack — it isn't realistic on 4GB RAM. Within that constraint, help me evaluate:
   - Batch speech-to-text (e.g., whisper.cpp or faster-whisper, tiny/base models) for voice notes — feasible as a background/on-demand job, not real-time
   - Local TTS (e.g., Piper) to have notes read back to me
   - Whether Joplin's own ecosystem (e.g., a Whisper-based Joplin plugin) already covers what I want with less custom infra than a bespoke Hugging Face/Ollama pipeline
   - Whether any of this is worth the added complexity at all, versus just doing it on a laptop/phone when needed
   
   Give a clear recommendation, not just a menu of options.

3. **Backup strategy to AWS S3.** I need one restorable backup, not a retention/versioning strategy (a couple of rotating snapshots is fine, no need for more). Cover:
   - What actually needs to be backed up — this depends on Joplin Server's `STORAGE_DRIVER` setting (attachments in Postgres vs. on the filesystem). Confirm which mode my setup uses before designing the backup, since it changes what gets backed up.
   - Postgres dump strategy (`pg_dump` from inside/outside the container) vs. volume-level backup
   - How secrets (Postgres password, SMTP credentials) should be handled in the backup — not stored unencrypted in the S3 bucket
   - Automation via cron (or a lightweight alternative) on the Pi, including how failures would surface to me (I don't want a silently broken backup)
   - AWS CLI is not yet installed on the Pi — note that as a setup step
   - How I'd actually test a restore, since an untested backup isn't a real backup

4. **Anything I'm missing.** Given this is a note store for critical work notes, flag anything in monitoring, disk health (SSD wear, free space alerts), or update strategy (Joplin/Postgres image updates) that I haven't asked about but should plan for.

Confirm your understanding of the current state (what's built vs. what's still a decision) before drafting `roadmap.md`.