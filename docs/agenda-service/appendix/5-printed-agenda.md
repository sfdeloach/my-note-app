This template is printed and handed to all elders at the meeting: standard 8-1/2" x 11" paper, portrait, greyscale (no color).

**What is normative here: the CSS and the DOM structure.** They are close to pixel-perfect and represent what the printed agenda should look like in its final form.

**What is not normative: the `<script>` block.** This was a one-off prototype with hard-coded data. In the service, the data comes from the parsed note plus `settings.json`, and rendering is a Go `html/template` with **no client-side JavaScript**. The script is retained only so the file still opens in a browser as a visual reference.

Notes on the prototype data below:

- The `elders` array duplicates `4-settings-and-data.md`, which is authoritative. It shows the 18 elders active at 2026-08-11.
- `ordered: true/false` stands in for `settings.listType`, which in the service is keyed by the exact h1 heading and defaults to `unordered`.
- The prototype's `isVisible` flags and nested sub-item lists have been removed. The Markdown convention can express neither, and both are out of scope.
- The `@page` rule, the screen-desk chrome, and the `@media print` block are shared with the Minutes and Action Items views. The font stack is **not** shared — this view uses Times New Roman, the other two use Cambria, and that difference is intentional.

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Session Meeting Agenda</title>
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
        font-family: "Times New Roman", Times, serif;
        font-size: 11pt;
        line-height: 1.35;
        color: #000;
        background: #fff;
        width: 7.7in; /* 8.5in paper - 0.4in margins */
        min-height: 10.2in; /* 11in paper - 0.4in margins */
        margin: 0.4in auto;
        box-shadow: 0 0 6px rgba(0, 0, 0, 0.25);
      }

      h1,
      h2,
      h3 {
        margin: 0;
        font-size: 11pt;
      }

      /* ---------- masthead ---------- */

      .masthead {
        display: grid;
        grid-template-columns: 1fr 1fr;
        align-items: start;
        background-color: #d0d0d0;
        padding: 18px 6px 32px;
        border-top: 1px solid #000;
        border-bottom: 3px double #000;
      }

      .masthead .church {
        font-size: 18pt;
        font-weight: bold;
      }

      .masthead h1 {
        font-size: 18pt;
      }

      .masthead .meeting-date {
        font-size: 18pt;
        font-weight: bold;
        margin: 0;
      }

      /* ---------- roll call ---------- */

      .roll-call {
        display: flex;
        gap: 1rem;
        padding: 12px 24px;
        border-bottom: 1px solid #000;
      }

      .roll-call h2 {
        flex: 1;
        font-weight: normal;
      }

      .roll-call ul {
        flex: 4;
        margin: 0;
        padding: 0;
        list-style: none;
        column-count: 3; /* fallback; overridden by elderColumns */
        column-gap: 1rem;
      }

      .roll-call li {
        break-inside: avoid;
      }

      /* Checkbox drawn in CSS so it renders identically everywhere,
         rather than relying on a glyph the serif face may not carry. */
      .roll-call li::before {
        content: "";
        display: inline-block;
        width: 0.5em;
        height: 0.5em;
        margin-right: 0.5em;
        border: 1px solid #000;
      }

      /* ---------- agenda ---------- */

      .agenda-title {
        background-color: #d0d0d0;
        padding: 0.5rem;
        border-bottom: 3px double #000;
        font-size: 16pt;
        font-weight: bold;
      }

      .agenda-body {
        padding: 1rem 3rem;
      }

      .agenda-body h3 {
        margin-top: 1rem;
        break-after: avoid;
      }

      /* Whether or not the note renders, the first heading sits flush
         against the .agenda-body padding. */
      #agenda-sections h3:first-child {
        margin-top: 0;
      }

      .agenda-body ol,
      .agenda-body ul {
        margin: 0;
      }

      .agenda-body li {
        break-inside: avoid;
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

        /* Browsers drop background fills when printing unless told otherwise. */
        .masthead,
        .agenda-title {
          -webkit-print-color-adjust: exact;
          print-color-adjust: exact;
        }
      }
    </style>
  </head>

  <body>
    <header class="masthead">
      <div class="church">Saint Andrew&rsquo;s Chapel</div>
      <div>
        <h1>Session Meeting Agenda</h1>
        <p class="meeting-date" id="meeting-date"></p>
      </div>
    </header>

    <section class="roll-call">
      <h2>Session Members:</h2>
      <ul id="roll-call-list"></ul>
    </section>

    <h2 class="agenda-title">Agenda</h2>

    <div class="agenda-body">
      <div id="agenda-sections"></div>
    </div>

    <script>
      "use strict";

      /* =========================================================
         PROTOTYPE DATA ONLY. In the service this comes from the
         parsed note plus settings.json, rendered server-side.
         ========================================================= */

      const AGENDA = {
        // ISO date, interpreted as UTC so it never shifts a day.
        meetingDate: "2026-08-11",

        elderColumns: 3,

        elders: [
          { firstName: "Burk", lastName: "Parsons" },
          { firstName: "Kevin", lastName: "Struyk" },
          { firstName: "Don", lastName: "Bailey" },
          { firstName: "Andrew", lastName: "Sarnicki" },
          { firstName: "Kevin", lastName: "Kennedy" },
          { firstName: "John", lastName: "Enslow" },
          { firstName: "Don", lastName: "McDade" },
          { firstName: "Chuck", lastName: "Micheals" },
          { firstName: "Michael", lastName: "Crotty" },
          { firstName: "Steve", lastName: "DeLoach", suffix: "Jr." },
          { firstName: "Lee", lastName: "Webb" },
          { firstName: "John", lastName: "Clendinen" },
          { firstName: "Alan", lastName: "Bird" },
          { firstName: "Tom", lastName: "Gibbons" },
          { firstName: "Robert", lastName: "Bisbing" },
          { firstName: "Dave", lastName: "Murray" },
          { firstName: "Bill", lastName: "Reisenweaver" },
          { firstName: "Ken", lastName: "Moody" },
        ],

        // ordered: true -> <ol>; false -> <ul>
        // Items are plain strings: the h2 titles, verbatim from the note.
        sections: [
          {
            heading: "Notes",
            ordered: false,
            items: ["Elders invited to come early to prayer at 4:30 PM"],
          },
          {
            heading: "Absences",
            ordered: true,
            items: ["Burk Parsons", "Dave Murray"],
          },
          {
            heading: "Reports & Updates",
            ordered: true,
            items: [
              "Opening Prayer",
              "Appoint a Moderator",
              "Membership Updates",
            ],
          },
          {
            heading: "New Business",
            ordered: true,
            items: [
              "Communication Liaison Committee",
              "Minister Resolution",
              "Bank Authorization",
            ],
          },
          {
            heading: "Reminders",
            ordered: false,
            items: ["September 1, 2026, Called Meeting"],
          },
        ],
      };

      /* ============================ renderer ============================ */

      function formatMeetingDate(isoDate) {
        return new Intl.DateTimeFormat("en-US", {
          month: "long",
          day: "numeric",
          year: "numeric",
          timeZone: "UTC",
        }).format(new Date(isoDate));
      }

      function byLastThenFirst(a, b) {
        return (
          a.lastName.localeCompare(b.lastName) ||
          a.firstName.localeCompare(b.firstName)
        );
      }

      function fullName(elder) {
        const base = `${elder.firstName} ${elder.lastName}`;
        return elder.suffix ? `${base} ${elder.suffix}` : base;
      }

      function buildList(items, ordered) {
        const list = document.createElement(ordered ? "ol" : "ul");

        for (const item of items) {
          const li = document.createElement("li");
          li.textContent = item;
          list.appendChild(li);
        }
        return list;
      }

      function renderSections(sections) {
        const container = document.getElementById("agenda-sections");

        for (const section of sections) {
          // A section with no items is omitted entirely, heading and all.
          if (!section.items || section.items.length === 0) continue;

          const heading = document.createElement("h3");
          heading.textContent = section.heading;
          container.append(
            heading,
            buildList(section.items, section.ordered !== false),
          );
        }
      }

      function renderRollCall(elders, columns) {
        const list = document.getElementById("roll-call-list");
        list.style.columnCount = columns;

        for (const elder of [...elders].sort(byLastThenFirst)) {
          const li = document.createElement("li");
          li.textContent = fullName(elder);
          list.appendChild(li);
        }
      }

      function render(agenda) {
        const dateText = formatMeetingDate(agenda.meetingDate);
        document.getElementById("meeting-date").textContent = dateText;
        document.title = `Session Meeting Agenda \u2014 ${dateText}`;

        renderRollCall(agenda.elders, agenda.elderColumns);
        renderSections(agenda.sections);
      }

      render(AGENDA);
    </script>
  </body>
</html>
```