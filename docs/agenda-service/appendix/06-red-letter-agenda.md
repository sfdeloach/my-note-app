This is copied and pasted into an email sent to the pastors. It is never printed.

This is **not** the printed agenda with red added. It is an independent, email-safe rendering that reuses the same section/item structure and the same `listType` settings. No masthead, no roll call, no `@page`, different type treatment. Don't build it by subclassing the agenda template.

Head matter: `{Type} Meeting Agenda`, then date and time, then location. No church name.

**The red is inlined on each span deliberately.** Most email clients strip `<style>` blocks on paste, which would silently drop the color in the one place it matters. The class is kept for the on-page view; the `style` attribute is what survives the paste.

Each `redLetter` value is wrapped in a literal `(Note: …)`. An item with no `redLetter` renders as a bare list item — see "Burk Parsons" and "Bank Authorization" below.

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Red Letter Agenda</title>
    <style>
      body {
        font-family: Cambria, Cochin, Georgia, Times, "Times New Roman", serif;
      }
      span.red-letter {
        color: firebrick;
      }
    </style>
  </head>
  <body>
    <main>
      <p>
        <strong>Stated Meeting Agenda</strong>
        <br />August 11, 2026, 5:00 PM
        <br />Classroom 7/8
      </p>
      <section>
        <p><strong>Notes</strong></p>
        <ul>
          <li>
            Elders invited to come early to prayer at 4:30 PM
            <span class="red-letter" style="color: firebrick;">
              (Note: Lee Webb would like "to lead the Session" in a time of
              prayer)
            </span>
          </li>
        </ul>
      </section>
      <section>
        <p><strong>Absences</strong></p>
        <ol>
          <li>Burk Parsons</li>
          <li>
            Dave Murray
            <span class="red-letter" style="color: firebrick;">
              (Note: Traveling, notified via email on 8/1)
            </span>
          </li>
        </ol>
      </section>
      <section>
        <p><strong>Reports &amp; Updates</strong></p>
        <ol>
          <li>Opening Prayer</li>
          <li>Appoint a Moderator</li>
          <li>Membership Updates</li>
        </ol>
      </section>
      <section>
        <p><strong>New Business</strong></p>
        <ol>
          <li>
            Communication Liaison Committee
            <span class="red-letter" style="color: firebrick;">
              (Note: A proposal that Kennedy, Struyk, and DeLoach form a liaison
              committee with Burk, his defense team, and Ligonier leadership.
              Regular meetings have been occurring on Wednesday morning.)
            </span>
          </li>
          <li>
            Minister Resolution
            <span class="red-letter" style="color: firebrick;">
              (Note: This is a follow-up to the motion passed at the July stated
              meeting to immediately recognize Andrew&rsquo;s membership at Saint
              Andrew&rsquo;s Chapel pending a successful congregation vote and
              with withdrawal from the North Texas Presbytery.)
            </span>
          </li>
          <li>Bank Authorization</li>
        </ol>
      </section>
      <section>
        <p><strong>Reminders</strong></p>
        <ul>
          <li>
            September 1, 2026, Called Meeting
            <span class="red-letter" style="color: firebrick;">
              (Note: This meeting was called for the purpose of meeting with
              Burk during July's stated meeting)
            </span>
          </li>
        </ul>
      </section>
    </main>
  </body>
</html>
```