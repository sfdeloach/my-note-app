This is the record of the meeting, printed on 8-1/2" x 11" paper.

Head matter: church name on the left, "Session Meeting Minutes" and the date on the right.

**The preamble is assembled from Metadata:**

```
The Saint Andrew's Session held a <strong>{type, lowercased} meeting</strong> in {location}. The meeting was called to order at {time}.
```

**The rosters are derived, not authored.** Present = elders active at the meeting date, minus those named in the `# Absences` section. Absent = those named. Both sorted last name then first. If nobody is absent, render `None`.

**Item titles and section headings do not appear.** Only `motion` and `comments` values render, in document order, and only for items that carry at least one. That is what makes it legitimate for the adjournment motion and the closing prayer to live under the "Bank Authorization" item in the master note — the title never surfaces.

- `motion` → `<p><strong>Motion</strong> {value}</p>`. The value already opens with the attribution parenthetical.
- `comments` → `<p>{value}</p>`, no label.

The `**REMOVE**` / `**ADD**` in the bank motion come from the note body and are rendered by the hand-rolled bold pass.

The `@page` rule, screen chrome, and `@media print` block are shared with the Agenda and Action Items views. The font stack is not — Cambria here, Times New Roman on the agenda, and that difference is intentional.

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Session Meeting Minutes</title>
    <style>
      /* ---------- page setup ---------- */

      @page {
        size: letter;
        /* 0.4in clears the unprintable hardware margin on most inkjets.
           0.25in fits laser printers but risks clipping the full-width
           rules at the paper edge. */
        margin: 0.4in;
      }

      /* Screen-only chrome: a gray desk so the page box is visible.
         Undone in the print block below. */
      html {
        background: #e8e8e8;
      }

      body {
        font-family: Cambria, Cochin, Georgia, Times, "Times New Roman", serif;
        font-size: 11pt;
        line-height: 1.35;
        color: #000;
        background: #fff;
        width: 7.7in; /* 8.5in paper - 0.4in margins */
        min-height: 10.2in; /* 11in paper - 0.4in margins */
        margin: 0.4in auto;
        box-shadow: 0 0 6px rgba(0, 0, 0, 0.25);
      }

      header {
        display: grid;
        grid-template-columns: 1fr 1fr;
        grid-template-rows: 1fr;
        margin-bottom: 2rem;
      }

      header h1,
      h2,
      h3 {
        margin: 0;
        padding: 0;
      }

      /* ---------- print ---------- */

      @media print {
        html {
          background: none;
        }

        /* Let the @page box define the width. Setting an explicit width
           that exactly equals the printable area invites rounding
           overflow and a stray blank page. */
        body {
          width: auto;
          min-height: 0;
          margin: 0;
          box-shadow: none;
        }
      }
    </style>
  </head>

  <body>
    <header>
      <div>
        <h2>Saint Andrew&rsquo;s Chapel</h2>
      </div>
      <div>
        <h2>Session Meeting Minutes</h2>
        <h2>August 11, 2026</h2>
      </div>
    </header>

    <p>
      The Saint Andrew&rsquo;s Session held a <strong>stated meeting</strong> in
      Classroom 7/8. The meeting was called to order at 5:00 PM.
    </p>

    <p>
      <strong> Session Members Present: </strong>
      Don Bailey, Alan Bird, Robert Bisbing, John Clendinen, Michael Crotty,
      Steve DeLoach Jr., John Enslow, Tom Gibbons, Kevin Kennedy, Don McDade,
      Chuck Micheals, Ken Moody, Bill Reisenweaver, Andrew Sarnicki, Kevin
      Struyk, Lee Webb
    </p>

    <p>
      <strong> Session Members Absent: </strong> Dave Murray, Burk Parsons
    </p>

    <p>Lee Webb opened the meeting in prayer.</p>

    <p>
      <strong>Motion</strong> (Crotty) to appoint Kevin Struyk as the moderator
      in Burk Parsons' absence, carried.
    </p>

    <p>
      <strong>Motion</strong> (DeLoach) and seconded that the Session authorize
      the following changes to the list of authorized signers on the Church's
      bank accounts, and direct the Director of Accounting to complete the
      necessary documentation with the bank: <strong>REMOVE</strong> Stephen
      Adams and Lee Webb, and <strong>ADD</strong> Rob Bisbing, Dave Murray, Ken
      Moody, Bill Reisenweaver, and Andrew Sarnicki
    </p>

    <p>
      <strong>Motion</strong> (Micheals) and seconded to adjourn the meeting.
    </p>

    <p>Bill Reisenweaver closed the meeting in prayer.</p>
  </body>
</html>
```