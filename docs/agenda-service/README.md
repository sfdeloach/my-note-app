# docs/agenda-service/

Prompt and design history for the **Agenda Service** (`agenda-service/` at
the repo root). Historical reference only — not loaded automatically, read
on demand when past reasoning is needed. Root `CLAUDE.md` has the current
state; `agenda-service/README.md` has the day-to-day authoring/dev guide.

## Convention: one numbered folder per feature

Each feature effort that adds to the service gets its own folder, named
`NN-<slug>/` with a zero-padded sequence number — same scheme as
`docs/conversations/`. The next feature is the next integer.

Inside a feature folder:

- **`prompt.md`** *(required)* — the expert prompt / build brief that
  specified the feature. This is the normative statement of intent for that
  effort.
- **`roadmap.md`** *(optional)* — a stage-by-stage breakdown plus a running
  "decided during implementation" log. Worth writing only for a feature
  large enough to span multiple work sessions; a small feature can skip it.
- Other feature-specific notes as needed.

When a new expert prompt arrives, create `NN-<slug>/prompt.md` and work
from there.

## `appendix/` — shared reference pool

Cross-feature supporting material, cited by any feature that needs it
rather than copied into a feature folder:

| File | What it is |
|---|---|
| `00-overview.md` | Original problem statement / workflow being replaced (non-normative history) |
| `01-example-note.md` | The canonical compliant master note — every authoring rule demonstrated in context |
| `02-scanner.md` | The original Markdown tree-scanner prototype the `parser` package started from |
| `03-scanner-result.md` | Exact expected parse tree for `01-example-note.md` |
| `04-settings-and-data.md` | Spec for `settings.json` (elder roster, per-section list types) |
| `05-printed-agenda.md` | Normative CSS/DOM spec — Agenda view |
| `06-red-letter-agenda.md` | Normative spec — Red-Letter Agenda view |
| `07-meeting-minutes.md` | Normative spec — Minutes view |
| `08-action-items-report.md` | Normative spec — Action Items view |
| `09-member-updates.md` | Normative spec — Minutes "Member Updates" section |

**`01-example-note.md`, `03-scanner-result.md`, and `04-settings-and-data.md`
are live Go test fixtures** — loaded by tests in
`agenda-service/{parser,views,reader,settings}/` via
`../../docs/agenda-service/appendix/…` paths. Edit them in lockstep with
those tests. New features extend this pool rather than fork it.

## What's here now

- **`01-initial-build/`** — the original bring-up: `prompt.md` (the build
  brief) + `roadmap.md` (7 stages, all `(done)`). The service is
  code-complete and live-deployed on the Pi.
- **`02-member-update-tables/`** — `prompt.md` for the Minutes "Member
  Updates" section (a `member-updates` fenced block → up to five tables).
  No `roadmap.md` — a one-to-two-session feature. Adds
  `appendix/09-member-updates.md`.
