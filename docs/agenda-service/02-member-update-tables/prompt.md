# Member Update Tables — build brief

## What I'm building and why

The Minutes view (`docs/agenda-service/appendix/07-meeting-minutes.md`) needs a
new section, **Member Updates**, inserted between the motion/comment paragraphs
and the "Attested by" attestation footer. It records every membership change
ratified at the meeting, in up to five tables:

1. New Members
2. Baptisms
3. Transfers
4. Removals
5. Deaths

Each table has its own column set and header titles — they are not uniform. Most
meetings touch only one or two of these, so **a table with no rows is not
rendered** (no empty tables, no empty headings). A meeting with no membership
changes at all shows no Member Updates section.

**This section appears in the Minutes view only.** It must never surface in the
Agenda, Red-Letter, or Action Items views. Whatever convention I adopt to record
the data in the master note must not disturb any existing authoring convention.

- Normative for the view being changed: `docs/agenda-service/appendix/07-meeting-minutes.md`.
- Canonical master note (the parser's contract, and a live test fixture):
  `docs/agenda-service/appendix/01-example-note.md`.
- **New** normative spec for the tables' DOM and CSS:
  `docs/agenda-service/appendix/09-member-updates.md` — to be authored as part of
  this effort, following the pattern of appendices 5–8 (a fenced ` ```html `
  golden document plus a "what is normative / what is not" preamble). I will
  supply the CSS I want; don't invent table styling before I do, however you may
  audit the CSS I provide to ensure it conforms to best practices.

---

## Confirmed system facts (don't rediscover)

Verified against the current code. Build against these.

- **Parser grammar** (`agenda-service/parser/parser.go:20–29`). Three line
  rules, each with a fixed depth:
  - `^# (.+)$` → `h1` section, depth 0
  - `^## (.+)$` → `item`, depth 1
  - `^- \*\*([A-Za-z]+):\*\* (\S.*)$` → key-value leaf, depth 2 (key is
    `[A-Za-z]+` only, value single-line)
  - `^- \*\*([A-Za-z]+):\*\*\s*$` → non-fatal warning, no block
  - `^>.*$` or a blank line → silently skipped
  - **anything else is a hard error that fails the whole note** — including a
    Markdown table row (`| … |`) and a code fence (` ``` `). This is why the new
    convention needs the parser change below, not just a new key.
- **The structural views only ever iterate top-level `h1` blocks**, skip
  `Metadata`, and skip any section with **zero `item` children**
  (`buildAgendaSections`, `buildRedLetterSections`; the zero-item behavior is
  locked by `emptySectionFixture` in `agenda-service/views/agenda_test.go`). The
  extractive views (Minutes motions/comments, Action Items) walk `item`
  children and emit nothing when there are none. **Consequence:** a `# Member
  Updates` h1 that contains no `##` items, plus a fenced block parsed as a
  *top-level, non-h1* block, is invisible to every existing view for free — no
  skip code anywhere. Do not add any.
- **Minutes template insert point:**
  `agenda-service/views/templates/minutes.tmpl` — between the `{{range .Entries}}`
  loop (currently lines 53–57) and the `<footer>` (currently lines 59–70). The
  Minutes and Action Items
  views share the Cambria font stack (`minutes.go`, `minutesActionItemsFontStack`)
  and the `printstyle` partial (`views/templates/print.tmpl`).
- **Table CSS to adapt:** `agenda-service/server/templates/list.tmpl` already has
  a `table`/`th`/`td` ruleset (`border-collapse`, bottom borders, top-aligned
  cells). Start from that, restyle for print / Cambria.
- **Dependencies:** stdlib only, plus `github.com/jackc/pgx/v5`. Inline Markdown
  is hand-rolled (`views/bold.go`). Keep it that way — no CSV library, no
  Markdown library.
- `docs/agenda-service/appendix/01-example-note.md` currently has an ad-hoc
  `## Membership Updates` item under `# Reports & Updates` (with a blockquote
  note). That placeholder is superseded by this feature — replace it with a real
  `member-updates` block when you update the fixture.

---

## How I'll record member updates in the master note

A single fenced block with the info string `member-updates`, under its own
`# Member Updates` h1, placed at the end of the note after `# Reminders`:

~~~text
# Member Updates

```member-updates
New Members
Name | Date Received | Notes
Jane Q. Doe | 2026-09-07 | Profession of faith
John & Mary Smith | 2026-09-07 | Transfer of letter, **PCA**

Deaths
Name | Date
Robert Roe | 2026-08-20
```
~~~

Rules for the block:

- **One `member-updates` block per note.** A second one is a per-note error.
- Inside, **stanzas separated by a blank line**. Each stanza is: line 1 = the
  table title, line 2 = the pipe-delimited header row, remaining lines =
  pipe-delimited data rows.
- **Pipe-delimited (`|`), cells trimmed.** No quoting rules, no escape — a
  literal `|` in a cell is simply unavailable, which is fine for names and
  dates. This keeps the hand-rolled split trivial.
- **Valid titles:** `New Members`, `Baptisms`, `Transfers`, `Removals`,
  `Deaths`. An unrecognized title is a per-note error (typo protection, same
  spirit as the Absences name check). Tables render in that canonical order
  regardless of authoring order.
- **Columns are whatever the header row declares.** The service does not
  hard-code a schema per table — it renders the header cells and the row cells
  as given. A data row whose cell count differs from its header row is a
  per-note error.
- A stanza with a title and header but **no data rows is omitted** from the
  output. No block, or a block with only omitted stanzas, means no Member
  Updates section.
- `**bold**` in a cell renders as `<strong>`, via the existing `views/bold.go`
  helper — same as every other authored value.

Why this doesn't collide with anything:

- Blockquotes stay exactly what they are — comments the parser drops.
- The fenced block is **new grammar**, not an overload of the key-value line or
  a heading.
- `# Member Updates` parses as an empty `h1` section, so the structural views
  skip it and the extractive views find no items under it. The block itself is a
  top-level non-`h1` node no view walker touches except the new Minutes builder.

---

## The parser change — one additive rule

`agenda-service/parser/parser.go` gains fenced-block capture, and nothing else
changes:

- On a line matching ` ^```(\S*)\s*$ ` while not already inside a fence, enter
  capture mode. Take every subsequent line **verbatim** — no rule matching, no
  "unrecognized line" error — until a closing ` ^```\s*$ `. Emit one **top-level
  (depth 0)** block carrying the info string and the raw body (exact `Block`
  shape is yours; it must not nest under an open `h1`/`item`).
- An **unterminated fence at EOF**: I lean toward a non-fatal `Warning` plus
  capture-to-EOF, consistent with how the parser already treats non-structural
  slips (the valueless-key rule). Argue for a hard error if you think it's
  better — flag it at checkpoint 1.
- Side benefit: fenced content anywhere in a note stops being fatal. Real
  legacy notes currently fail to parse for exactly this reason (see the
  open-items tail of `docs/agenda-service/01-initial-build/roadmap.md`).

**Fixtures move in lockstep**, same as commit `771465d` did for the attestation
footer:

- Add a `member-updates` block to `docs/agenda-service/appendix/01-example-note.md`
  (replacing the ad-hoc `## Membership Updates` item).
- Regenerate the expected tree in
  `docs/agenda-service/appendix/03-scanner-result.md` so
  `agenda-service/parser/parser_test.go` still passes.

---

## The Minutes view change

- `agenda-service/views/minutes.go`: add `MemberUpdates []MemberUpdateTable` to
  `MinutesModel`, where a table is roughly
  `{ Title string; Headers []string; Rows [][]template.HTML }`. Add
  `buildMemberUpdates(tree)` and call it from `BuildMinutesModel`. It finds the
  top-level `member-updates` block, splits stanzas, validates titles, checks
  each row's cell count against its header, skips empty stanzas, orders the
  tables canonically, and runs each cell through `Bold`.
- `agenda-service/views/templates/minutes.tmpl`: a
  `{{range .MemberUpdates}} … <table> … {{end}}` block between the `.Entries`
  loop and `<footer>`, plus a table ruleset in the `<style>` block adapted from
  `server/templates/list.tmpl`. Exact heading levels and DOM come from
  `appendix/09-member-updates.md`.
- **No other view changes. No `reader/`, `server/`, or `settings/` changes. No
  new dependency.**

---

## Roadmap overview

No standalone `roadmap.md`. Per `docs/agenda-service/README.md`, a roadmap is
"worth writing only for a feature large enough to span multiple work sessions,"
and this is one — two at most. The staging, if it helps:

1. **Parser rule + fixtures.** Add fenced-block capture; update
   `01-example-note.md` and regenerate `03-scanner-result.md`.
   *Verify:* `go test ./agenda-service/parser/...` passes; feeding every real
   note through the parser produces no new hard errors.
2. **Minutes wiring + golden spec.** `MemberUpdateTable` model, `buildMemberUpdates`,
   the template block and CSS; author `appendix/09-member-updates.md` with my
   CSS.
   *Verify:* `go test ./agenda-service/views/...` passes with new model and
   render-substring assertions; the rendered Minutes for the example note match
   `appendix/09`.
3. **Docs + closeout.** Extend the authoring section of
   `agenda-service/README.md`; add the pointer in `07-meeting-minutes.md`;
   update `CLAUDE.md` if warranted; add the next numbered
   `docs/conversations/` entry.
   *Verify:* `go build ./agenda-service/... && go vet ./agenda-service/...`
   clean; live smoke test against the Pi shows the section in the Minutes and
   nowhere else.

---

## Out of scope

- **Member Updates render in the Minutes view only** — never the Agenda,
  Red-Letter, or Action Items views.
- **Not a new view.** No new route, no `viewDispatch` entry, no separate
  download artifact. It is a section inside the existing Minutes output.
- **No writes to Joplin.** The fenced block is read-only input like everything
  else.
- **No new dependency.** Stanza and pipe splitting are hand-rolled.
- **No backfill.** Past meeting notes are not retrofitted with `member-updates`
  blocks as part of this work.
- **No fixed per-table schema in code**, no totals, no roll-ups, no
  year-to-date membership statistics — the tables render exactly as authored.
- **No styling knobs.** The CSS is fixed in the template, like every other
  view.

---

## Deliverables

1. `agenda-service/parser/parser.go` — fenced-block capture rule; new unit tests
   in `parser_test.go` (terminated, unterminated, info string, a fence
   containing lines that would otherwise be unrecognized).
2. `docs/agenda-service/appendix/01-example-note.md` +
   `docs/agenda-service/appendix/03-scanner-result.md` — a `member-updates`
   block added and the expected tree regenerated, in lockstep.
3. `agenda-service/views/minutes.go` — `MemberUpdateTable`, `MemberUpdates`
   model field, `buildMemberUpdates`.
4. `agenda-service/views/templates/minutes.tmpl` — the `{{range .MemberUpdates}}`
   block and table CSS.
5. `docs/agenda-service/appendix/09-member-updates.md` — normative DOM/CSS spec
   (golden fenced HTML + "what is normative" preamble); a one-line pointer added
   to `docs/agenda-service/appendix/07-meeting-minutes.md`.
6. `agenda-service/README.md` — the authoring-convention section extended with
   the `member-updates` block format, noting it renders in the Minutes only.
7. `agenda-service/views/minutes_test.go` — model assertions in
   `TestBuildMinutesModel_ExampleNote` and render-substring assertions in
   `TestRenderMinutes_ExampleNote`; a hard-error case for an unknown table title
   and for a row/header cell-count mismatch.
8. `CLAUDE.md` updated if warranted, and the next numbered `docs/conversations/`
   entry summarizing what was built and decided.

---

## Checkpoints — stop and show me at each

1. **Parser first.** Implement the fenced-block rule, show me it captures the
   `member-updates` block in the updated `01-example-note.md` and reproduces the
   regenerated `03-scanner-result.md` exactly, and confirm no existing note now
   hard-errors that didn't before. Note your call on the unterminated-fence
   case. Stop there.
2. **Authoring format + golden spec.** Show me the finalized block format and a
   draft `appendix/09-member-updates.md` — I'll hand you the CSS — before you
   wire the template. Stop there.
3. Then the Minutes model, builder, template, and tests.

Scope and sequencing beyond checkpoint 2 are yours to decide.

---

Reference: `docs/agenda-service/appendix/09-member-updates.md` (to be authored).
