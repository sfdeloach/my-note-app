# Agenda Service — build brief

## What I'm building and why

I run a self-hosted Joplin stack (Joplin Server + Postgres) via Docker Compose on a Raspberry Pi 5, reachable only over WireGuard or my trusted LAN. Each month I prepare a stated Session meeting agenda for my church. I author one **master note in Joplin**, and I want a small **Go service** that reads it and renders four derived views on demand, served over the tunnel.

The master note is also my private copy: anything I want kept to myself I write in a **blockquote**, which the parser ignores entirely. So there is no separate "private" view — the master in Joplin *is* it, and nothing private ever reaches a rendered output.

Background on the workflow this replaces is in `docs/agenda-service/appendix/0-overview.md`. That file is **non-normative history** — where it disagrees with this brief, this brief wins.

---

## Confirmed system facts (already verified against the live DB — don't rediscover)

I've inspected the running database. Build against these; do not treat any of this as unknown:

- **Nothing is encrypted at rest.** All items show `jop_encryption_applied = 0`. Direct Postgres reads are valid.
- **Table `items`** holds everything. Relevant columns: `content` (`bytea`), `jop_id` (`varchar(32)`), `jop_parent_id` (`varchar(32)`), `jop_type` (`integer`: **1 = note, 2 = notebook/folder**), `jop_encryption_applied` (`integer`), `owner_id`, `title`.
- **`content` is a UTF-8 JSON object**, not the legacy newline-delimited Joplin format. Decode the bytea as UTF-8, then JSON-parse. The envelope carries `title`, `body` (the markdown I authored), and `markup_language` (**1 = Markdown**).
- **Notebooks.** Meeting masters live in a Joplin notebook titled **"Session Meetings"**. Inside it is a sub-notebook titled **"Archived"** holding older meeting notes, which must continue to render on demand. "Archived" is currently empty.

Read path: fetch note row → decode `content` bytea as UTF-8 → JSON-parse envelope → take `body` string → hand to the parser.

**Do not ask me for the notebook name or for HTML templates.** The notebook is "Session Meetings"; the templates are appendices 5–8.

### Verify once at startup (fatal on failure)

These are the only things a Joplin version bump could plausibly break, so assert them at boot and fail fast with a clear message rather than degrading quietly:

- The columns above exist with the stated types.
- "Session Meetings" resolves to exactly one `jop_type = 2` row.
- "Archived" resolves to exactly one `jop_type = 2` row **whose `jop_parent_id` is the "Session Meetings" id**. Do not match "Archived" globally by title — it's a plausible notebook name elsewhere in my Joplin.
- Every note in both notebooks has `jop_encryption_applied = 0`.

Two things I have *not* verified, so check them during the first build and tell me what you find:

- Whether `items` retains soft-deleted rows in my Joplin Server version. If it does, exclude them from note listings.
- `owner_id` — this is a single-user instance, so I expect one value. Confirm it, and if there's more than one, filter on mine.

`markup_language` lives inside the JSON envelope, not in a column, so it can't be asserted in SQL at startup. Check it **per note at read time** and treat a violation as a per-note error (below).

### Startup-fatal vs. per-note error

These two rules are deliberately different and must not be collapsed:

- **Startup checks are fatal.** If the schema or the notebooks aren't what I described, refuse to start.
- **A single bad note must never take down the container.** `markup_language = 2`, a malformed body, or a title/metadata disagreement is an error attached to *that note* — surfaced on that note's page and in its list entry — while every other note keeps serving.

---

## Architecture

**Direct Postgres, read-only.** Joplin *Server* has no Data API (that belongs to the desktop/CLI client, which I deliberately don't run here), so reading Postgres is the real path, not a shortcut. The cost is coupling to Joplin's internal schema, so contain it:

- Put **all** Joplin-schema and JSON-envelope knowledge behind a single `reader` package with a narrow interface returning clean Go values. Nothing outside `reader` knows how Joplin stores data. If a Joplin update changes the schema, there's one file to fix.
- Add a line to the update/restore ritual (`UPDATING.md`) noting that Joplin version bumps should smoke-test this service.

The interface, roughly:

```go
type Scope int // Current, Archived

type Note struct {
    ID    string
    Title string
    Date  string // YYYY-MM-DD, parsed from Title
    Type  string // e.g. "Stated", parsed from Title
}

type NoteBody struct {
    Note
    Body string
}

ListNotes(scope Scope) ([]Note, error)
GetNote(id string) (NoteBody, error)
```

`GetNote` must **reject any id that isn't a note in one of the two notebooks** — the service should never render an arbitrary row from my Joplin database because someone guessed a `jop_id`.

**Dedicated read-only DB role.** Do not reuse the `joplin` credentials. Create a Postgres role with **SELECT-only** access scoped to `items`. Rationale: a bug in this app must never be able to modify or delete my notes. Show me the `CREATE ROLE`/`GRANT` and the tradeoffs — I'm solid in Go but weaker on DB permissions and networking, so explain the *why*. Its password goes in `.env` (real, never printed/committed) with a placeholder in `.env.example`.

---

## Three layers

1. **`reader`** — Postgres + JSON envelope (above).
2. **`parser`** — my existing tree scanner, hardened (below).
3. **`views`** — one renderer per view (below).

Views are pure over their inputs but they need the settings file as well as the tree, so the signature is `func(tree []*parser.Block, cfg settings.Settings) (ViewModel, error)`, with `html/template` doing the rendering. **All four views are server-rendered Go templates. No client-side JavaScript in any of them.**

---

## Note title convention

Master notes are titled `YYYY-MM-DD <Type> Meeting` — e.g. `2026-08-11 Stated Meeting`, `2026-10-15 Called Meeting`.

The title duplicates the Metadata `date` and `type`, which means they can drift when I copy last month's note forward and forget to rename it. So: **parse the title into date and type, and cross-check both against the Metadata section. A disagreement is a per-note error** naming both values. That catches the copy-forward mistake at exactly the moment I'd make it.

A title that doesn't match the pattern at all is also a per-note error.

Listing sorts by **title descending** — the ISO prefix makes that chronological without parsing every note body to build the list page.

---

## Authoring convention (the parser's contract)

Derived from my example note, `docs/agenda-service/appendix/1-example-note.md`. The parser depends only on this convention, never on any existing note content.

- The note body is Markdown. A **blockquote (`>`) is a comment** — the parser ignores it completely. This is how I keep private notes in the master.
- The body is divided into **h1 sections (`# `)**: a special **`# Metadata`** section plus five body sections — **Notes, Absences, Reports & Updates, New Business, Reminders**. The set of body sections is open; a sixth would be treated as a normal body section.
- Each agenda item is an **h2 (`## `)** under a section.
- Item content is **key-value lines**: `- **key:** value`. Regex: `^- \*\*([A-Za-z]+):\*\* (\S.*)$`.
- **Values are single-line.** A soft-wrapped value is a parse error, not a continuation — I'd rather the error tell me than have the service silently drop half a motion.
- Values are **trimmed** of leading and trailing whitespace.
- **Keys can repeat under one item** (e.g., an item with two `motion` lines). Preserve all, in document order.
- A key line **with no value** (`^- \*\*([A-Za-z]+):\*\*\s*$`) is **skipped with a non-fatal warning**, not an error. This rule doesn't exist in the current scanner and must be added — see the parser section.
- Keys in use: `date`, `time`, `location`, `type` (Metadata only); and `redLetter`, `motion`, `comments`, `actionItem` (items). The set is open — capture any key; views select the ones they care about.

**Metadata is special-cased.** `# Metadata` contains **exactly one h2**, whose title is ignored (it's "Info" in my example). Zero h2s, or more than one, is a parse error. Its key-values (`date` in ISO 8601, `time`, `location`, `type`) populate each view's head matter. Metadata is **never rendered as a body section**. Every other h1 is a normal body section.

**Parse errors** (each names the note and the line number):

- h3 or deeper.
- Prose or any other unrecognized line outside a blockquote.
- A key-value line before any h2, or an h2 before any h1 (the existing skip-level guard already covers this).
- A multi-line / continued value.

---

## The parser — start from my scanner, then harden

`docs/agenda-service/appendix/2-scanner.md` holds my existing prototype in a fenced Go block. It already builds the correct tree: h1 → h2 → key-value, using a depth/stack discipline with a skip-level guard. **Keep that logic and the `Block{Key, Content, Children}` shape.** Note that the prototype declares `package service` and takes an `*os.File` — both are artifacts of the prototype. The target is **`package parser`**.

Required changes to make it service-grade:

- **Return errors, never `log.Fatalf`.** Change the signature to accept the body string (or an `io.Reader`) and return `([]*Block, []Warning, error)`.
- **Add the valueless-key rule.** Right now `- **motion:**` matches neither the keyword rule nor `safeSkipRegex`, so it falls through to the unrecognized-line error. Add the rule above, matched before the error branch, emitting a warning.
- **Trim values.**
- Otherwise preserve behavior: blockquotes and blank lines skipped; anything else unrecognized is a parse error for that note.

`docs/agenda-service/appendix/3-scanner-result.md` is the **exact expected output** for `1-example-note.md`. Treat it as the parser's test fixture.

---

## Settings and data

`docs/agenda-service/appendix/4-settings-and-data.md` describes a JSON file holding what doesn't belong in the notes because it changes infrequently and would clutter them: the elder roster and per-section list types.

- Lives at `agenda-service/config/settings.json`, path supplied by `AGENDA_SETTINGS_PATH`, bind-mounted read-only into the container.
- **Validated at startup — fatal on failure.**
- **Re-read and re-parsed per request**, so a roster edit takes effect without restarting the container. At my volume the cost is irrelevant.
- If a per-request read or parse fails, serve a **clear error page naming the file and the JSON error**. Do not fall back to a last-good copy — a broken roster riding along silently in a document of record is worse than a visible failure.

**`listType` is keyed by the exact h1 title** — `"Notes"`, `"Absences"`, `"Reports & Updates"`, `"New Business"`, `"Reminders"`. A section not present in the map **defaults to unordered**.

**Elder active/inactive** is computed against the Metadata `date`. `removedAt` is the elder's **last active day**, so a meeting held on that date still includes him. `elderClass` is captured but unused by any view.

Elder names carry an optional `suffix` field, rendered as `First Last Suffix` — e.g. "Steve DeLoach Jr." **No comma.** The Minutes rosters are comma-separated inline lists, where "Steve DeLoach, Jr." would read as two people. One form everywhere is worth more than the extra comma.

---

## Inline Markdown in values

Some values contain `**bold**` (the bank motion's `**REMOVE**` / `**ADD**`), and the views render it as `<strong>`. I want **zero dependencies**, so hand-roll it:

1. HTML-escape the value first.
2. Then substitute paired `**…**` into `<strong>…</strong>`.

Escaping before substituting is what keeps an author from injecting markup. An unmatched `**` is left as literal asterisks — it'll be visible in the output, which is signal enough.

Bold is the only inline construct in evidence. Don't implement anything else.

---

## The four views

Two are *structural* (full section/item skeleton); two are *extractive* (walk the tree and conditionally emit items carrying a target key).

**There is no shared header across the four views** — each has its own head matter, specified below. The three print views *do* share `@page`, screen-desk chrome, and the `@media print` block; factor those into one stylesheet.

**Do not unify the font stacks.** The agenda uses `"Times New Roman", Times, serif`; the minutes and action items use `Cambria, Cochin, Georgia, Times, "Times New Roman", serif`. This is intentional — these documents have always differed, and the point of automating them is that no reader notices anything changed.

**Dates** render as `January 2, 2006`, parsed from the ISO string **as UTC** so a timezone can never shift the day.

**Empty sections are omitted entirely**, heading and all, in both structural views.

**Sub-items and per-meeting visibility flags are dropped.** Appendix 5's prototype JS supports nested `{ text, items: [...] }` lists and an `isVisible` flag; the Markdown convention can express neither, and I don't want them back.

### 1. Agenda (print, 8.5×11). Structural.

The hardcopy handed to the elders. Reference: `appendix/5-printed-agenda.md`.

- **Head matter:** church name on the left, "Session Meeting Agenda" with the date beneath it on the right. `type` is not used.
- **Roll-call band** beneath the masthead: every elder active at the meeting date, sorted last name then first, in `elderColumns` columns, each with a CSS-drawn checkbox.
- Then each body section (h1) with its item titles (h2), as an `<ol>` or `<ul>` per `listType`.
- **No key-values render.**

### 2. Red-letter agenda / Pastor's copy (screen only — NOT printed). Structural.

Reference: `appendix/6-red-letter-agenda.md`.

This is **not** the Agenda view with red added. It's an **independent, email-safe rendering** that happens to reuse the same section/item structure and the same `listType` settings. No masthead, no roll call, no `@page`, different type treatment. Don't try to subclass the agenda template.

- **Head matter:** "{Type} Meeting Agenda" (type cased as authored), then date and time, then location. No church name.
- Each item's `redLetter` value(s) render after the item title, wrapped in a literal `(Note: …)`, in red.
- **The red must survive a paste into an email client.** Most clients strip `<style>` blocks on paste, so **inline the color on each span** in the copied fragment. Keep the class for the on-page view if you like, but the inline attribute is what does the work.
- Provide a copy affordance — select-to-copy, or a button using the Clipboard API to write `text/html`.
- **Do not build any email-sending feature.** A copyable block is the whole deliverable.

### 3. Minutes (print, 8.5×11). Extractive.

Reference: `appendix/7-meeting-minutes.md`.

- **Head matter:** church name on the left, "Session Meeting Minutes" and the date on the right.
- **Preamble**, assembled from Metadata:
  `The Saint Andrew's Session held a <strong>{type, lowercased} meeting</strong> in {location}. The meeting was called to order at {time}.`
- **Rosters**, derived from the settings roster and the Absences section:
  - Absent = the elders named by the h2 titles under `# Absences`.
  - Present = elders active at the meeting date, minus the absent.
  - **Matching:** case-insensitive exact match of the h2 title against `"{firstName} {lastName}"` (suffix excluded from the match key). **No match is a hard error** naming the unmatched title — a typo'd name in a document of record should stop the render, not silently produce a wrong roster.
  - Both lists sorted last name then first. If nobody is absent, render `None`.
- Then, **in document order**, the items carrying `motion` and/or `comments` children, rendering those values in document order within each item. Items with neither are omitted.
  - `motion` → `<p><strong>Motion</strong> {value}</p>`. The value already begins with the attribution parenthetical.
  - `comments` → `<p>{value}</p>`, no label.
- **Section headers are omitted, and so are the h2 item titles.** That's what makes it legitimate for the adjournment motion and closing prayer to live under the "Bank Authorization" item in the master — the title never surfaces.
- Repeated `motion` lines under one item all render.

### 4. Action Items (print, 8.5×11). Extractive.

Reference: `appendix/8-action-items-report.md`.

- **Head matter:** "Saint Andrew's Chapel Session Meeting Action Items" on one line, date in bold beneath, both centered. `type` is not used.
- Then a flat `<ol>` — one `<li>` per h2 item that has an `actionItem` child, containing the item title and one `<p><strong>Action Item:</strong> {value}</p>` per `actionItem` child. Items without one are omitted.

---

## HTTP surface — keep it tiny

Small Go binary, standard-library-leaning, minimal dependencies:

- `GET /` — list current master notes (excludes archived).
- `GET /archived` — list notes from the "Archived" sub-notebook.
- `GET /view?id=<note>&view=agenda|redletter|minutes|actionitems` — render the chosen view (printable HTML for the three print views; screen HTML with the copy affordance for red-letter).
- Optionally `GET /download?...` for the three print views, with `Content-Disposition` filenames `session-agenda-2026-08-11.html`, `session-minutes-2026-08-11.html`, `session-action-items-2026-08-11.html`. Red-letter has no download.

Binds inside the compose network, reached over WireGuard/LAN — same trust model as the rest of the stack. No public exposure.

**Environment variables:** `AGENDA_DB_HOST`, `AGENDA_DB_PORT`, `AGENDA_DB_NAME`, `AGENDA_DB_USER`, `AGENDA_DB_PASSWORD`, `AGENDA_SETTINGS_PATH`, `AGENDA_LISTEN_ADDR`.

---

## Respect these existing project decisions — don't re-litigate

Per the repo's `CLAUDE.md`:

- **No reverse proxy / TLS** — WireGuard encrypts the tunnel, the LAN is trusted.
- **Manual, deliberate updates; no `latest` tags** — pin your base image.
- **Storage mode stays `Database`** — don't touch Joplin's storage config.
- This service moves `my-note-app` off "pure maintenance config," so **update `CLAUDE.md`** (there's now app code: the `agenda-service/` subdir, its build/run commands, the new env vars) and **add the next numbered `docs/conversations/` entry** summarizing what we built and decided.
- **Communication style:** explain the *why* and tradeoffs for anything touching the DB role, ports, or the tunnel — I'm weaker there than in Go.

---

## Out of scope

- No writing to or migrating Joplin's tables — ever. Read-only, always.
- **No email-sending.** The red-letter view is a copyable block, nothing more.
- No headless Joplin client, no reverse proxy, no auth beyond the tunnel, no Prometheus/Grafana.
- No client-side JavaScript beyond the red-letter copy button.

---

## Deliverables

1. `agenda-service/` — Go source: `reader` (all Joplin coupling), `parser` (hardened), `settings`, `views` (four renderers), HTTP layer.
2. `agenda-service/config/settings.json` seeded from appendix 4.
3. `Dockerfile` for the service.
4. A service entry added to the existing `docker-compose.yml`, joining the current network, with the settings file bind-mounted read-only.
5. The read-only Postgres role (SQL + where I run it), with its credential placeholdered in `.env.example`.
6. Updated `CLAUDE.md` and a new numbered `docs/conversations/` entry.
7. A short `agenda-service/README.md` covering the authoring convention (so I remember how to write masters), the note title convention, and how to add a new view.

---

## Checkpoints — stop and show me at each

1. **Parser first.** Implement `parser` against `1-example-note.md` and show me it reproduces `3-scanner-result.md` exactly. Stop there.
2. **Then the DB role.** Show me the `CREATE ROLE`/`GRANT` SQL with the reasoning and tradeoffs, before touching `docker-compose.yml`. Stop there.
3. Then views, then the HTTP layer.

Scope and sequencing beyond those checkpoints are yours to decide.