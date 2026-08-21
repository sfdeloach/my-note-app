This tracks what needs to be done after the meeting, printed on 8-1/2" x 11" paper.

Head matter: "Saint Andrew's Chapel Session Meeting Action Items" on one line, date in bold beneath, both centered. The meeting `type` is not used.

Then a flat `<ol>` — one `<li>` per h2 item that has at least one `actionItem` child, containing the item title and one labeled paragraph per `actionItem`. Items without one are omitted, and so are section headings.

The `@page` rule, screen chrome, and `@media print` block are shared with the Agenda and Minutes views.

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Session Meeting Action Items</title>
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
        display: flex;
        flex-direction: column;
        align-items: center;
        margin-bottom: 2rem;
      }

      p {
        margin: 0 auto;
      }

      li {
        margin-bottom: 1rem;
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
      <p>Saint Andrew&rsquo;s Chapel Session Meeting Action Items</p>
      <p><strong> August 11, 2026 </strong></p>
    </header>

    <ol>
      <li>
        <p>Bank Authorization</p>
        <p>
          <strong> Action Item: </strong>
          DeLoach to collect signatures on the provided bank paperwork from the
          new elders and return to Stassia.
        </p>
      </li>
    </ol>
  </body>
</html>
```