# Roadmap: Agenda Service

This roadmap breaks the Agenda Service build brief
(`docs/agenda-service/initial-prompt.md`, plus the appendices in
`docs/agenda-service/appendix/`) into stages sized to complete one at a
time in future conversations. Each stage has a **Goal**, a rough list of
**Tasks**, and a **Verify** step to confirm it worked before moving on —
same format as the top-level `docs/roadmap.md`.

The brief itself sets two hard stops (its "Checkpoints" section): after the
parser, and after the DB role proposal. Everything after that is this
roadmap's own sequencing, not the brief's.

**Status key:** not started / in progress / done / blocked.

---

## Decisions already made (context for all stages — don't re-litigate)

Pulled from the brief; see `initial-prompt.md` for full reasoning:

- **Direct Postgres, read-only, always.** No Data API exists on Joplin
  Server, so this is the real path, not a shortcut. No headless Joplin
  client, ever.
- **All Joplin-schema and JSON-envelope knowledge lives in one place**: the
  `reader` package. Nothing else knows how Joplin stores data.
- **Dedicated read-only DB role**, not the `joplin` credentials — a bug in
  this app must never be able to touch the notes.
- **Zero dependencies for inline Markdown.** Bold (`**…**`) is hand-rolled:
  escape first, then substitute.
- **`settings.json` is re-read and re-parsed per request**, never cached,
  never falls back to a last-good copy on error.
- **All four views are server-rendered `html/template`.** No client-side
  JavaScript anywhere except the Red-Letter copy button.
- **No reverse proxy, no TLS, no auth beyond the tunnel** — same trust model
  as the rest of the stack (`CLAUDE.md`). Revisit only if that project-level
  decision changes.
- **Manual, deliberate image versions** — pin the service's own Dockerfile
  base image; no `latest`.
- **Startup-fatal vs. per-note error are deliberately different** and must
  never be collapsed: schema/notebook/role problems refuse to start the
  service; a single bad note (bad `markup_language`, malformed body,
  title/Metadata disagreement) is an error scoped to that note while
  everything else keeps serving.

### Corrections to the brief

The brief's HTTP section refers to "the existing `docker-compose.yml`." The
actual file in this repo is **`compose.yml`**. Use the real name in Stage 6.

Confirmed via live DB recon in Stage 2:
- **`items` has no `title` or `deleted_time` column.** Both — along with
  `body` and `markup_language` — live only inside the JSON envelope in the
  `content` bytea column, never as row-level columns. The row-level `name`
  column is a synthetic filename (`{jop_id}.md`), not the title. This
  means note listing/sorting can't be done with plain SQL on `title`; the
  `reader` package decodes+JSON-parses `content` for every candidate row.
- **Soft-deleted rows are retained**, marked by a nonzero
  `content.deleted_time` inside the envelope (confirmed via a live
  example) — not a row-level flag.
- **`owner_id` is single-valued**, as expected — confirmed via
  `SELECT DISTINCT owner_id FROM items`.

Decided during implementation in Stage 3 (not corrections, just choices the
brief left open):
- **Dates stay plain ISO (`YYYY-MM-DD`) strings throughout**, not
  `time.Time` — matches `reader.Note.Date`'s existing convention, and the
  fixed-width zero-padded format makes lexicographic string comparison
  correct for the active-at-date logic. `time.Parse` is used exactly once,
  inside `settings.Load`'s validation step, only to reject a
  syntactically-valid-JSON-but-impossible date at startup.
- **`Settings` is a plain value type**, not a pointer — unlike
  `reader.Reader`, it holds no live resource (it's rebuilt from disk on
  every call, never cached), matching the brief's `cfg settings.Settings`
  by-value view signature.
- **Validation beyond the brief's explicit text** (all fatal, at
  `settings.Load`): `elderColumns` must be positive; each elder's
  `firstName`/`lastName`/`elderClass` must be non-empty. `elderClass`'s
  *value* is deliberately left unrestricted (not locked to
  teaching/ruling) since no view reads it yet.

Decided during implementation in Stage 4 (not corrections, just choices the
brief left open):
- **Metadata-section extraction lives in `views`, not `reader`.** The
  brief's view signature is `(tree []*parser.Block, cfg settings.Settings)`
  — views never receive the already-validated `reader.Note`, so they need
  their own read of the `# Metadata` section's `date`/`time`/`location`/
  `type` values. `reader/envelope.go`'s `findMetadata` does a similar walk
  but only extracts `date`/`type` for the title cross-check and is
  unexported; rather than export DB-package internals for a non-DB reason,
  `views` has its own small unexported `extractMetadata`, reused as-is by
  every later view that needs it.
- **`churchName` is a private const in `agenda.go`** for now, not a shared
  file — only Agenda uses it in Stage 4. Promote it to a shared location
  when Stage 5's Action Items needs it too, per "build once, reuse the
  first time a second consumer needs it."
- **The shared print stylesheet takes `FontStack` as `template.CSS`, not
  `string`.** `html/template`'s contextual CSS escaper mangles a plain
  string's quotes/commas when interpolated inside `<style>`; `template.CSS`
  is the documented escape hatch for a well-formed CSS fragment the code
  itself controls.
- **Agenda and Red-Letter's section/item walks are two small separate
  functions** (`buildAgendaSections`, `buildRedLetterSections`), not one
  generic walk — they produce different item shapes (`[]string` vs.
  `[]RedLetterItem`), and a shared generic helper wasn't worth it for two
  call sites.
- **Bold-inline substitution uses `regexp.MustCompile(`\*\*(.+?)\*\*`)`**
  (non-greedy, applied after `template.HTMLEscapeString`) rather than
  manual delimiter-pair scanning — simpler, and naturally leaves any
  unmatched `**` literal without special-casing it.

Decided during implementation in Stage 5 (not corrections, just choices the
brief left open):
- **The entry-extraction walk is its own small function per view**
  (`buildMinutesEntries`, `buildActionItemEntries`), not a variant of
  `buildAgendaSections`/`buildRedLetterSections`. Those two filter by
  *item presence* and always surface a heading/title; Minutes' and Action
  Items' walks filter by *descendant key* and Minutes never surfaces a
  heading or item title at all — different enough shapes that forcing a
  shared generic would cost more than two small functions, consistent with
  Stage 4's same call on Agenda vs. Red-Letter.
- **Present/Absent are pre-joined `string` fields**, not `[]string` —
  unlike `AgendaModel.Elders` (rendered as a `<ul>`, one `<li>` per name),
  Minutes' rosters render as a single comma-separated run of text inside
  one `<p>`, so `strings.Join` in `BuildMinutesModel` is simpler than
  teaching the template to join. `Absent` is the literal string `"None"`
  when nobody's absent, matching the appendix wording directly.
- **Absent-elder exclusion from Present uses a `FirstName+"\x00"+LastName`
  map key, not `settings.Elder` struct equality** — `Elder.RemovedAt` is a
  `*string`, so two independently-obtained copies of the same elder (one
  from `ActiveElders`, one matched off the `# Absences` section) aren't
  `==`-comparable even when they represent the same person.
- **`minutesActionItemsFontStack` lives in `minutes.go`**, reused as-is by
  `actionitems.go` — the two views share the identical Cambria stack (per
  appendix 7/8), so it's declared once where Minutes needed it first and
  referenced, not redeclared, by Action Items.
- **No `Absences` section in the note is not an error** — `absentElders`
  returns `(nil, nil)` rather than failing, since an all-hands meeting
  legitimately has nothing under `# Absences`. Only a *name that doesn't
  match any elder* is a hard error, per the brief.

Decided during implementation in Stage 6 (not corrections, just choices the
brief left open):
- **Entry point is `agenda-service/main.go` at the module root**, not a
  `cmd/` subdirectory — this module only ever builds one binary, so the
  extra nesting wasn't earning its keep.
- **A new `server` package** holds routing, HTTP handlers, and the
  note-listing template — kept separate from `views`, which is scoped to
  the four print/screen views, not note listing. `main.go` stays pure
  wiring: env vars → `pgxpool.Pool` → `reader.New` → one startup-only
  `settings.Load` (to fail fast on a broken settings file, same as the
  brief's fatal-at-startup rule) → `server.New` → `http.Server`.
- **Router is stdlib `net/http`'s Go 1.22+ pattern-based `ServeMux`**
  (`"GET /view"`-style patterns, `"GET /{$}"` to match `/` exactly rather
  than as a catch-all). No router dependency added — `go.mod` had none,
  and this matches the project's existing zero-dependency approach to
  Markdown parsing.
- **The note listing (`/` and `/archived`) is one handler parameterized by
  `reader.Scope`**, rendering a small embedded `server/templates/list.tmpl`
  (same `embed.FS` pattern as `views/render.go`). Each note's four view
  links and three download links are precomputed in Go
  (`noteViews`), not built inside the template, so URL-escaping the note
  id happens in one place. A note with `Err` set renders as an inline
  error message instead of links — as of Stage 6, this is every single
  real note (see "Open items"), which the listing handles correctly by
  construction rather than as a special case.
- **Download filename convention**: `session-<view>-<date>.html` (e.g.
  `session-agenda-2026-08-11.html`) for Agenda/Minutes/Action Items;
  Red-Letter has no `FilenamePrefix` and `/download?view=redletter` is
  rejected outright (400) rather than left to fail some other way.
- **Every failure path renders into a `bytes.Buffer` before writing to the
  `http.ResponseWriter`**, in both `handleList` and `handleView` — so a
  render error partway through (a malformed note, Minutes' unmatched-name
  hard error) still becomes a clean error response instead of a
  half-written 200. Error responses are deliberately graduated: 400 for a
  bad/unknown query param, 404 for `reader.ErrNoteNotFound`, 422 for a
  per-note content problem (`*reader.NoteError`, or a `Build*Model`/
  `Render*` failure) with a message naming the note and the problem, and
  500 (generic body, real error logged server-side only) for anything
  that might carry DB or filesystem detail.
- **Dockerfile**: multi-stage, `golang:1.25.14-trixie` builder (matching
  `go.mod`'s `go 1.25.0` and the project's Debian/trixie image family) →
  `debian:trixie-slim` runtime (chosen over a distroless base — keeps a
  shell for `docker exec -it` debugging, consistent with how the
  `database` container is already worked with; the tradeoff is a larger
  image and more attack surface than shell-less, acceptable given the
  tunnel/LAN-only trust model). `CGO_ENABLED=0` static build (pgx is pure
  Go), runs as an explicitly created non-root user. Verified locally:
  image builds, container starts and connects to Postgres over the
  compose network, and runs with a working shell as a non-root uid.
- **`compose.yml` publishes one host port** (`8080:8080`, `AGENDA_LISTEN_ADDR`
  internally `:8080`), the same shape as `joplin-server`'s `22300:22300` —
  needed because WireGuard tunnels to the Pi's host, not into the Docker
  bridge network, so a service with no published port would be
  unreachable from outside the host entirely. The file's top-of-file "no
  reverse proxy" comment was updated (not left stale): agenda-service
  *is* the second service its old wording said to revisit for, and the
  reasoning was reconfirmed, not reversed.
- **All `AGENDA_*` config is pulled through `.env`/`.env.example`**, even
  values effectively fixed by topology (`AGENDA_DB_HOST=database`,
  `AGENDA_SETTINGS_PATH=/etc/agenda-service/settings.json`) — matches how
  `joplin-server`'s own config is already 100% `.env`-sourced in this
  file, rather than introducing a mixed hardcoded/templated convention.

### Shared building blocks — build once, reuse

These span multiple stages. Build each the first time it's needed and reuse
it rather than re-deriving it:

- **Bold-inline renderer** (escape-then-substitute `**…**` → `<strong>`) —
  first needed by Red-Letter (Stage 4), reused by Minutes' bank motion
  (Stage 5).
- **Print stylesheet** (`@page`, screen-desk chrome, `@media print`) shared
  by Agenda, Minutes, and Action Items (not Red-Letter, which is
  screen-only) — first built in Stage 4 with Agenda, reused in Stage 5.
  Font stacks are deliberately **not** unified (Times New Roman for Agenda;
  Cambria/Cochin/Georgia for Minutes and Action Items).
- **Elder sort** (last name, then first name) and **date formatting**
  (`January 2, 2006`, ISO string parsed as UTC) — used across Agenda,
  Red-Letter, and Minutes.

### Deliverables → stages

The brief lists 7 deliverables. Mapping, for tracking:

1. `agenda-service/` Go source (`reader`, `parser`, `settings`, `views`,
   HTTP) — Stages 1–3, 6.
2. `agenda-service/config/settings.json` seeded from appendix 4 — Stage 3.
3. `Dockerfile` — Stage 6.
4. Service entry in `compose.yml` — Stage 6.
5. Read-only Postgres role (SQL + where run), credential in `.env.example`
   — Stage 2 (proposed), applied Stage 2/3 boundary, wired in Stage 6.
6. Updated `CLAUDE.md` + new `docs/conversations/` entry — Stage 7.
7. `agenda-service/README.md` — Stage 7.

---

## Stage 1 — Scaffolding + parser (checkpoint 1) (done)

**Goal**: a standalone, hardened `parser` package that reproduces the
brief's exact test fixture.

**Tasks**:
- Create the `agenda-service/` Go module and the package layout
  (`reader/`, `parser/`, `settings/`, `views/`, plus a `cmd/` or root
  `main.go` for later).
- Port the scanner prototype (`appendix/2-scanner.md`) into
  `package parser`: keep the depth/stack discipline, the skip-level guard,
  and the `Block{Key, Content, Children}` shape.
- Change the signature to accept the body string (or `io.Reader`) and
  return `([]*Block, []Warning, error)` — no `log.Fatalf` anywhere.
- Add the valueless-key rule (`^- \*\*([A-Za-z]+):\*\*\s*$`), matched
  *before* the unrecognized-line error branch, emitting a warning instead
  of an error.
- Trim values.
- Preserve everything else: blockquotes and blank lines skipped; anything
  else unrecognized is still a parse error, with note + line number.

**Verify**: running the parser against `appendix/1-example-note.md`
reproduces `appendix/3-scanner-result.md`'s tree exactly, and produces
exactly 3 warnings (for the valueless `motion`/`comments`/`actionItem`
lines under "Minister Resolution") — assert on key names and warning
count, not exact line numbers, per the fixture's own note. **Stop and show
this before continuing** (brief's checkpoint 1).

---

## Stage 2 — DB recon + `reader` package + DB role proposal (checkpoint 2) (done)

**Goal**: resolve the two open unknowns against the live database, build
the `reader` package against the confirmed schema, and propose the
dedicated read-only role.

**Tasks**:
- Query the live DB (existing credentials) to resolve the two things the
  brief flags as unverified:
  - Does `items` retain soft-deleted rows in this Joplin Server version? If
    so, note listings must exclude them.
  - What values does `owner_id` actually take? Confirm single-user
    (expected); if more than one, note that queries must filter to mine.
  - Report findings back before or alongside the reader implementation —
    the brief asks to be told what's found here.
- Implement the `reader` package interface from the brief:
  `Scope`, `Note{ID, Title, Date, Type}`, `NoteBody{Note, Body}`,
  `ListNotes(scope) ([]Note, error)`, `GetNote(id) (NoteBody, error)`.
- `GetNote` must reject any id that isn't a note row inside "Session
  Meetings" or its "Archived" child — never render an arbitrary `jop_id`.
- Read path: fetch row → decode `content` bytea as UTF-8 → JSON-parse →
  take `body` → hand to `parser`.
- Startup-fatal verification (assert once at boot, refuse to start on
  failure): the `items` columns/types described in the brief exist;
  "Session Meetings" resolves to exactly one `jop_type = 2` row; "Archived"
  resolves to exactly one `jop_type = 2` row whose `jop_parent_id` is
  Session Meetings' id (not matched globally by title); every note in both
  notebooks has `jop_encryption_applied = 0`.
- Per-note checks (error scoped to that note, others keep serving):
  `markup_language` must be 1; parse the title against
  `YYYY-MM-DD <Type> Meeting` and cross-check both date and type against
  the Metadata section — a mismatch, or a title that doesn't match the
  pattern at all, is a per-note error naming both values.
- Listing sort: title descending (ISO prefix makes this chronological
  without parsing every body).
- Draft the `CREATE ROLE` / `GRANT` SQL: a role with **SELECT-only** access
  scoped to `items`, separate from the `joplin` credentials. Write up the
  reasoning and tradeoffs (why not reuse `joplin`'s role, what SELECT-only
  does and doesn't protect against, connection/network implications) —
  the user is solid in Go but weaker on DB permissions, so this needs the
  *why*, not just the SQL.

**Verify**: reader package startup checks pass against the live DB;
`GetNote` on an id outside the two notebooks is rejected; a deliberately
mismatched title/Metadata pair produces a per-note error, not a crash.
**Stop and show the `CREATE ROLE`/`GRANT` SQL with its reasoning before
touching `compose.yml`** (brief's checkpoint 2) — role creation/application
itself can follow approval within this stage, since the reader needs a
working role to run against; wiring the credential into `compose.yml`
waits for Stage 6.

---

## Stage 3 — Settings package (done)

**Goal**: load and validate `settings.json`, and implement the elder
roster logic every view depends on.

**Tasks**:
- Load from `AGENDA_SETTINGS_PATH`; validate at startup, fatal on failure.
- Re-read and re-parse on every request (not cached); on a per-request
  read/parse failure, serve a clear error page naming the file and the
  JSON error — no fallback to a last-good copy.
- `listType` lookup keyed by exact h1 title; a section absent from the map
  defaults to unordered.
- Elder active/inactive-at-date: active if
  `activeAt <= meetingDate AND (removedAt undefined OR meetingDate <= removedAt)`
  — `removedAt` is the elder's *last active day*, so a meeting held that
  day still includes him.
- Name formatting: `First Last[ Suffix]`, no comma, ever (this is what
  keeps "Steve DeLoach Jr." from reading as two people in the Minutes'
  comma-separated rosters).
- `elderClass` is parsed/stored but not read by any view — don't wire it
  anywhere yet.

**Verify**: `settings.json` seeded from `appendix/4-settings-and-data.md`
loads cleanly; a deliberately malformed JSON file produces the clear
per-request error page, not a crash or a stale render; elder
active/inactive at `2026-08-11` matches the brief's stated 18-active
result (confirming Zima and Villavicencio, both past `removedAt`, are
excluded).

---

## Stage 4 — Structural views: Agenda + Red-Letter Agenda (done)

**Goal**: the two "full skeleton" views, plus the shared building blocks
they introduce (bold-inline renderer, print stylesheet).

**Tasks**:
- Implement the bold-inline renderer: HTML-escape the value, then
  substitute paired `**…**` into `<strong>…</strong>`; an unmatched `**` is
  left literal.
- Build the shared section/item-walk logic: skip Metadata (never a body
  section), omit empty sections entirely (heading and all), render each
  h1's h2 items as `<ol>`/`<ul>` per `listType`.
- **Agenda** (`appendix/5-printed-agenda.md` is normative for CSS/DOM,
  print-only, 8.5×11): masthead (church name left; "Session Meeting
  Agenda" + date right; `type` unused); roll-call band of elders active at
  the meeting date, sorted last-then-first, in `elderColumns` CSS columns
  with CSS-drawn checkboxes; body sections with item titles only — no
  key-values render. Introduce the shared print stylesheet here.
- **Red-Letter Agenda** (`appendix/6-red-letter-agenda.md`, screen-only,
  **never printed**, independent template — not a subclass of Agenda): head
  matter is `{Type} Meeting Agenda` (type cased as authored) + date/time +
  location, no church name; each item's `redLetter` value(s) render after
  the title as a literal `(Note: …)` with the red color **inlined on the
  span** (paste survival — most email clients strip `<style>` blocks); a
  copy affordance (select-to-copy or a Clipboard API button writing
  `text/html`). No email-sending feature — the copyable block is the whole
  deliverable.

**Verify**: both views render against the example note + seeded settings;
Agenda matches appendix 5's DOM/CSS structure; Red-Letter's copied HTML
fragment (paste-tested into an actual email client, or inspected for the
inline `style` attribute) retains red without a `<style>` block; an empty
section (hypothetically) is omitted entirely in both.

---

## Stage 5 — Extractive views: Minutes + Action Items (done)

**Goal**: the two "walk and extract" views, reusing Stage 4's shared
building blocks.

**Tasks**:
- Build the shared item-extraction walk: collect, in document order, items
  carrying a target key (or keys); items with none are omitted; section
  headers and h2 item titles never surface in either view.
- **Minutes** (`appendix/7-meeting-minutes.md`, print, shares Stage 4's
  print stylesheet, Cambria font stack): head matter (church name left,
  "Session Meeting Minutes" + date right); preamble assembled from
  Metadata (`type` lowercased, `location`, `time`); Present/Absent rosters
  — Absent = h2 titles under `# Absences`, matched case-insensitively
  against `"{firstName} {lastName}"` (suffix excluded from the match key),
  a non-match is a **hard error** naming the unmatched title; Present =
  active-at-date roster minus Absent; both sorted last-then-first; `None`
  if nobody absent. Then, in document order, items carrying `motion`
  and/or `comments`: `motion` → `<p><strong>Motion</strong> {value}</p>`
  (repeats all render), `comments` → `<p>{value}</p>`. The bank motion's
  `**REMOVE**`/`**ADD**` runs through Stage 4's bold-inline renderer.
- **Action Items** (`appendix/8-action-items-report.md`, print, shares the
  print stylesheet, Cambria font stack): centered head matter ("Saint
  Andrew's Chapel Session Meeting Action Items" + bold date, `type`
  unused); flat `<ol>`, one `<li>` per h2 item carrying at least one
  `actionItem`, with the item title and one
  `<p><strong>Action Item:</strong> {value}</p>` per `actionItem` child.

**Verify**: both views render against the example note; Minutes' rendered
Present/Absent lists match the brief's worked example (16 present, 2
absent — Murray, Parsons — at the example note's date); a deliberately
misspelled name under `# Absences` produces the hard error, not a silently
wrong roster; the adjournment motion and closing prayer correctly surface
under "Bank Authorization" with no visible item title or section heading.

---

## Stage 6 — HTTP layer + Dockerfile + compose integration (checkpoint 3, cont'd) (done)

**Goal**: a tiny, tunnel-only HTTP surface, containerized and wired into
the existing stack.

**Tasks**:
- Routes: `GET /` (current notes, excludes archived), `GET /archived`,
  `GET /view?id=<note>&view=agenda|redletter|minutes|actionitems`
  (printable HTML for the three print views, screen HTML + copy affordance
  for red-letter), optionally `GET /download?...` for the three print
  views with `Content-Disposition` filenames
  (`session-agenda-YYYY-MM-DD.html`, etc. — no download for red-letter).
- Env vars: `AGENDA_DB_HOST`, `AGENDA_DB_PORT`, `AGENDA_DB_NAME`,
  `AGENDA_DB_USER`, `AGENDA_DB_PASSWORD`, `AGENDA_SETTINGS_PATH`,
  `AGENDA_LISTEN_ADDR`.
- `Dockerfile` for the service, pinned base image (no `latest`).
- Add the service to `compose.yml` (the actual filename — see correction
  above), joining the existing network, `settings.json` bind-mounted
  read-only, no host port exposure beyond what the tunnel/LAN trust model
  already covers.
- Add the Stage 2 role's real password to `.env` (never committed) and its
  placeholder to `.env.example`.

**Verify**: `docker compose up -d` starts the new service cleanly
alongside Joplin/Postgres; all four views reachable over the tunnel/LAN by
note id; a bad/missing `id` or `view` param fails cleanly, not with a
stack trace or a 500 that leaks DB detail; download filenames match the
spec for the three print views.

Confirmed live on the Pi: `docker compose up -d` (via `--build
agenda-service`) started it cleanly alongside `database`/`joplin-server`
(`docker compose ps` shows it `Up`, publishing `8080:8080`); `GET /` and
`GET /archived` both return 200, with every one of the 36 current real
notes showing as an inline error row (none are yet authored in the
compliant convention — expected, see "Open items"); `GET /view` with a
missing/unknown param returns 400, with a nonexistent id returns 404, and
`GET /download?view=redletter` is rejected with 400; the `server` package's
integration tests (gated on `AGENDA_TEST_DB_*`, same pattern as
`reader/reader_test.go`) pass against the live database, including the
per-note-error path against the same real bad-title note
`reader/reader_test.go` already uses. **Not yet verified**: a full 200
render of any of the four views against a real note, since no real note
currently complies with the authoring convention (same gap
`reader/reader_test.go`'s `TestGetNote` already documents) — add that
check once a compliant note exists.

---

## Stage 7 — Docs & wrap-up (not started)

**Goal**: close the loop on the project's own documentation conventions
now that this repo has app code, not just Compose/maintenance config.

**Tasks**:
- Update root `CLAUDE.md`: this service moves `my-note-app` off "pure
  maintenance config" — document `agenda-service/`, its build/run
  commands, and the new env vars, alongside the existing Compose commands.
- Add the next numbered `docs/conversations/` entry summarizing what was
  built and decided across Stages 1–7.
- Write `agenda-service/README.md`: the authoring convention (so the
  master-note format is remembered), the note title convention, and how to
  add a new view.
- Add a line to `UPDATING.md` noting that Joplin version bumps should
  smoke-test this service (per the brief's coupling concern).
- End-to-end smoke test against a real (not example) note, across all four
  views.

**Verify**: `CLAUDE.md`, `UPDATING.md`, and the new conversation entry all
reflect the shipped state; `agenda-service/README.md` alone is enough to
remind a future self how to author a compliant master note; a real
month's meeting note renders correctly in all four views.

---

## Open items to track (not blocking any stage yet)

- ~~Confirm during Stage 2: soft-deleted row retention in `items`, and
  `owner_id` cardinality~~ — resolved in Stage 2, see "Corrections to the
  brief" below.
- Implementation-level decisions the brief explicitly defers (Postgres
  driver choice, Go module path, HTTP router, template file layout, exact
  Dockerfile base image/tag) are each stage's own call when reached — not
  resolved here. Stage 2 chose `github.com/jackc/pgx/v5` (native, not
  `database/sql`) for the Postgres driver.
- **None of the 34 real meeting notes currently comply with the h1/h2/
  key-value authoring convention** — they predate it (legacy freeform
  headers, markdown tables, a missing colon here and there). Confirmed via
  live recon in Stage 2. This isn't a bug — the per-note-error design
  exists for exactly this — but expect every real note to show a per-note
  error in Stages 4–6's views until notes are authored (or rewritten) in
  the new convention. Two titles also don't match the `YYYY-MM-DD <Type>
  Meeting` pattern at all: `2025-11-08 Called Teleconference` and
  `2025-09-16 Called Meeting/DZ Court`. "Template" and "Notes Converter
  Prompt" (non-meeting scratch notes) also live inside the Session
  Meetings notebook and will show the same title-pattern error.
