# Agenda Service

A small Go HTTP service that renders four printable/screen views (Agenda,
Red-Letter Agenda, Minutes, Action Items) directly from Session Meeting notes
stored in the Joplin Server database this stack runs. It reads Postgres
directly, read-only, through a dedicated `agenda_reader` role — there's no
Joplin Data API to go through, and this service never writes to `items`.
Like the rest of this stack, it has no reverse proxy, no TLS, and no auth of
its own: it's reachable only over the WireGuard tunnel or trusted home LAN,
same trust model as Joplin Server itself (see the root `CLAUDE.md`).

Full design history — why direct Postgres, why no headless Joplin client,
the per-stage decisions — lives in `docs/agenda-service/roadmap.md` and the
original brief, `docs/agenda-service/initial-prompt.md`. This file only
covers what's needed day to day: how to author a note this service can
render, and how to add a new view.

## Authoring a compliant meeting note

The service only understands one Markdown convention. A note that doesn't
follow it renders as a per-note error in the listing — the rest of the
notebook keeps serving normally.

**Title**: `YYYY-MM-DD <Type> Meeting`, e.g. `2026-08-11 Stated Meeting`.
The service cross-checks this against the Metadata section below and
errors if they disagree.

**Structure**: one `# Metadata` section, plus body sections — each a
top-level `# ` heading. The current body sections (matching
`config/settings.json`'s `listType` keys) are `# Notes`, `# Absences`,
`# Reports & Updates`, `# New Business`, and `# Reminders`.

**Metadata section**: exactly one `## ` heading (its title is ignored)
containing four key-values:

```markdown
# Metadata

## Info

- **date:** 2026-08-11
- **time:** 5:00 PM
- **location:** Classroom 7/8
- **type:** Stated
```

`date` is ISO `YYYY-MM-DD`; `type` must match the title's `<Type>` exactly.
Metadata is never rendered as a body section — its values populate every
view's head matter.

**Sections and items**: within a body section, each agenda item/entry is a
`## ` heading. Key-value data nested under an item uses:

```markdown
- **key:** value
```

(regex: `^- \*\*([A-Za-z]+):\*\* (\S.*)$`). Values are single-line. A
key line with no value (`- **key:**`) is skipped with a warning, not an
error — handy for a key you haven't filled in yet. Blockquotes (`> ...`)
and blank lines are always ignored, so blockquotes are free to use as
inline comments anywhere.

**Keys the views actually read:**

| Key | View | Effect |
|---|---|---|
| `redLetter` | Red-Letter Agenda | Renders as `(Note: …)` after the item title, in inline red |
| `motion` | Minutes | `<p><strong>Motion</strong> {value}</p>`, one per occurrence |
| `comments` | Minutes | `<p>{value}</p>` |
| `actionItem` | Action Items | `<p><strong>Action Item:</strong> {value}</p>`, one per occurrence |

An item with none of these keys still renders (title only) in Agenda and
Red-Letter, but is omitted entirely from Minutes' and Action Items'
extraction walks — those two only ever show items carrying their key.

**`# Absences`**: each `## ` heading under it is an elder's name, matched
case-insensitively against `"{firstName} {lastName}"` from
`config/settings.json` (suffix, e.g. "Jr.", is excluded from the match). A
name that doesn't match any elder is a **hard error** for the whole note —
this is deliberate, to catch typos rather than silently dropping someone
from the roster. No `# Absences` section at all is fine (not every meeting
has absences).

**Full worked example**: `docs/agenda-service/appendix/1-example-note.md`
is the brief's reference note — every rule above is demonstrated there,
including the valueless-key and Absences-error cases in context.

Bold text (`**…**`) inside any value renders as `<strong>` in every view
that shows it (e.g. the bank-authorization motion's `**REMOVE**`/`**ADD**`
in the example note).

## Adding a new view

Each view is a `Build*Model`/`Render*` pair in `views/`, e.g.
`views/agenda.go`'s `BuildAgendaModel` + `RenderAgenda`. `BuildAgendaModel`
walks the parsed `[]*parser.Block` tree and a `settings.Settings` value into
a template-ready struct; `Render*` executes the matching template. Templates
live in `views/templates/*.tmpl` and are wired in via `views/render.go`'s
`embed.FS`.

To add one:

1. Add `views/<name>.go` with `Build<Name>Model` and `Render<Name>` (model
   struct, `tmpl.ExecuteTemplate`). Reuse existing helpers where they fit:
   `Bold` (bold.go), `FormatDate` (dateformat.go), the shared print
   stylesheet (currently declared in `agenda.go`, reused by Minutes/Action
   Items), the shared elder-sort/active-at logic in `settings/elders.go`.
2. Add `views/templates/<name>.tmpl`.
3. Register it in `server/view.go`'s `viewDispatch` slice: a `viewDef` with
   the query-param name, display label, the `Render*` func, and (if it
   should be downloadable) `Downloadable: true` plus a `FilenamePrefix` —
   downloads are named `{FilenamePrefix}-{note-date}.html` by
   `server/view.go`'s `handleView`. Leaving `Downloadable` false rejects
   `/download` for that view outright (this is how Red-Letter, which has
   nothing to download, is handled).
4. No routing changes needed — `GET /view` and `GET /download` already
   dispatch through `viewDispatch` by param name.

## Build, run, test

```
go build ./...
go vet ./...
go test ./...                     # unit tests only, no DB needed
```

The `reader` and `server` packages also have integration tests gated on
`AGENDA_TEST_DB_*` env vars (host/port/name/user/password) — they skip
cleanly when those aren't set, and are read-only against whatever database
they're pointed at. Point them at the same read-only `agenda_reader`
credentials `.env` already has for the running service; since Postgres has
no host-published port (`compose.yml`), run them from a container attached
to the compose network, with `AGENDA_TEST_DB_HOST=database`:

```
docker run --rm --network my-note-app_my-note-app-network \
  -e AGENDA_TEST_DB_HOST=database -e AGENDA_TEST_DB_PORT=5432 \
  -e AGENDA_TEST_DB_NAME=joplin -e AGENDA_TEST_DB_USER=agenda_reader \
  -e AGENDA_TEST_DB_PASSWORD=<from .env> \
  -v "$(pwd)/..:/repo:ro" -w /repo/agenda-service \
  golang:1.25.14-trixie go test ./...
```

To run the service itself:

```
docker compose up -d --build agenda-service
docker compose logs -f agenda-service
```

## Env vars

All `AGENDA_*` vars are documented with placeholders in the root
`.env.example`:

| Var | Purpose |
|---|---|
| `AGENDA_DB_HOST` | Postgres host — `database` (Compose DNS name) in production |
| `AGENDA_DB_PORT` | Postgres port — `5432` |
| `AGENDA_DB_NAME` | Database name — `joplin` |
| `AGENDA_DB_USER` | The read-only role — `agenda_reader` |
| `AGENDA_DB_PASSWORD` | That role's password (real value only in `.env`, gitignored) |
| `AGENDA_SETTINGS_PATH` | In-container path to the bind-mounted `settings.json` |
| `AGENDA_LISTEN_ADDR` | HTTP listen address, e.g. `:8080` |
