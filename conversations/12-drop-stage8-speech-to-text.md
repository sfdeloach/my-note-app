# 12 — Drop Stage 8 (speech-to-text)

**Request:** Scope Stage 8 ("Speech-to-text for voice notes") from
`roadmap.md`.

**Decision:** drop Stage 8 entirely, before any implementation. Not
renumbered — kept visible in `roadmap.md` as "(dropped)" with rationale, the
same pattern used for "(done)" stages, so this isn't re-litigated later
without new information.

**What research surfaced:**
- Joplin **Server** (the sync backend this repo runs) has no Data API of
  its own — only Joplin Desktop/Mobile/CLI clients expose one, via the Web
  Clipper REST API on port 41184.
- Writing a transcript back into a note safely therefore isn't as simple as
  querying Postgres directly: notes are versioned sync items, and editing
  rows directly would bypass Joplin's own sync/versioning logic. The real
  path would be running a headless Joplin client (e.g. "Joplin Terminal")
  on the Pi as a sync client, just to get API access for the script.
- Separately, whisper.cpp itself checks out fine on this hardware (tiny/base
  models run near-real-time on a Pi 5, well within the 4GB budget as an
  on-demand job) — the API-access problem was the real complexity, not the
  transcription itself.
- The Joplin plugin ecosystem already has pieces of this (a built-in
  whisper.cpp-based "Voice Typing" live-dictation feature, a community
  "Dictate" plugin, an "Audio Transcriber" plugin using cloud APIs), but
  nothing that fits a self-hosted-server + on-demand-batch-job shape
  cleanly.

**Why dropped (the actual deciding factor):** on reconsideration, the use
case itself is thin. Native OS/phone dictation already covers live
note-taking, which is the common case. Transcribing an already-recorded
long audio file is infrequent enough to handle manually in a separate app
when it comes up, rather than building and maintaining a pipeline
(including a headless Joplin client) to support it. The one thing given up
is local/private transcription of existing audio — cloud dictation services
send audio off-device — judged acceptable given how rarely this need
arises.

**Done:**
- `roadmap.md`: Stage 8 heading changed to "(dropped)", Goal/Tasks/Verify
  replaced with a short rationale paragraph; "Decisions already made" AI
  features bullet narrowed to text-to-speech only.
- `CLAUDE.md`: stage list in "Where progress lives" marks speech-to-text as
  "(dropped)"; "Architecture decisions already made" AI features bullet
  narrowed to text-to-speech (Stage 9) with a note on why speech-to-text was
  dropped.
- This file: new conversation log entry.

**Next:** scope Stage 9 (text-to-speech) in a follow-up turn.
