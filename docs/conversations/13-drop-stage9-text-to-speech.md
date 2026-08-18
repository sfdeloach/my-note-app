# 13 — Drop Stage 9 (text-to-speech)

**Request:** Scope Stage 9 ("Text-to-speech for notes") from `roadmap.md`.

**Decision:** drop Stage 9 entirely, before any implementation — same
pattern as Stage 8 (see `conversations/12-drop-stage8-speech-to-text.md`).
This resolves the entire roadmap: all 9 stages are now done or deliberately
dropped.

**The real ask, clarified up front:** the roadmap's "set up Piper" framing
undersold what was actually wanted — a one-tap play button inside the
Joplin phone app, browse to any note and listen while driving or doing
chores. Worth researching before scoping a script, since that's a much
bigger ask than an offline batch converter.

**What research surfaced:**
- Joplin's own ecosystem has no read-aloud feature at all — an open,
  unresolved forum request dating back to 2019
  (discourse.joplinapp.org/t/text-to-speech/2355).
- Both Android ("Select to Speak") and iOS ("Speak Screen", two-finger swipe
  down) are built into the OS, work inside *any* app including Joplin, are
  fully offline-capable, and require zero engineering effort. This already
  covers the real underlying need — listening to a note hands-free — today,
  for free.

**Options considered for the "true play button" vision:**
1. A custom Joplin plugin plus a small always-listening HTTP endpoint on
   the Pi for the plugin to call. Conflicts directly with this project's
   existing "no resident services" and "no reverse proxy" decisions — a
   meaningfully bigger project than anything built so far.
2. Pre-generate audio attachments via a batch job. Hits the same
   headless-Joplin-client-for-API-access wall Stage 8 did (Joplin Server has
   no Data API), and isn't spontaneous — notes must be pre-converted before
   driving, not browsed live.
3. A standalone script with no Joplin integration — feed it note text, get
   an MP3, move it to the phone by hand. Small and matches "on-demand batch
   job," but isn't the in-app "play" experience originally pictured.

**Why dropped:** given OS-level read-aloud already solves the actual need
for free today, none of the three options above earn their added
complexity. Consistent with the Stage 8 call: don't build infrastructure
for a need a native platform capability already covers.

**Done:**
- `roadmap.md`: Stage 9 heading changed to "(dropped)", Goal/Tasks/Verify
  replaced with a rationale paragraph; "Decisions already made" AI features
  bullet rewritten to cover both dropped stages; added a closing note that
  the roadmap is now fully resolved.
- `CLAUDE.md`: stage list in "Where progress lives" marks text-to-speech as
  "(dropped)" too; "Architecture decisions already made" AI features bullet
  rewritten to cover both dropped stages and why.
- This file: new conversation log entry.

**Next:** none — roadmap complete. Any future work here (e.g. Postgres
minor version bump noted in `UPDATING.md`) is ad hoc maintenance, not a new
roadmap stage.
