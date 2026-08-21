> Blockquotes are used like comments in code. The scanner will ignore them. For the purpose of this example meeting note, they will be used to explain the convention used to take notes. In practice, anytime I want to add text to any part of my notes that I want the scanner to ignore, I will place the text in a blockquote.
> The note's Joplin title follows the pattern `YYYY-MM-DD <Type> Meeting` — this note is titled "2026-08-11 Stated Meeting". The service cross-checks that title against the Metadata section below and errors if they disagree.
> All notes shall be divided into a `# Metadata` section plus five body sections, each an h1 header.
> key-value convention: `- **key:** value`
> regex for key-values: `^- \*\*([A-Za-z]+):\*\* (\S.*)$`
> Values are single-line. A key line with no value is skipped with a warning.
> The `# Metadata` section holds exactly one h2 — its title is ignored — containing four key-value pairs: date (ISO 8601 YYYY-MM-DD format), time, location, and type (all strings). This section is never rendered as a body section; its values populate the head matter of each view.

# Metadata

## Info

- **date:** 2026-08-11
- **time:** 5:00 PM
- **location:** Classroom 7/8
- **type:** Stated

# Notes

## Elders invited to come early to prayer at 4:30 PM

- **redLetter:** Lee Webb would like "to lead the Session" in a time of prayer

# Absences

> An absence with no redLetter renders as a bare name.

## Burk Parsons

## Dave Murray

- **redLetter:** Traveling, notified via email on 8/1

# Reports & Updates

## Opening Prayer

- **comments:** Lee Webb opened the meeting in prayer.

## Appoint a Moderator

- **motion:** (Crotty) to appoint Kevin Struyk as the moderator in Burk Parsons' absence, carried.

## Membership Updates

> make sure the Smiths are included in the update

# New Business

> All agenda items under New Business are indicated by an h2 heading.

## Communication Liaison Committee

- **redLetter:** A proposal that Kennedy, Struyk, and DeLoach form a liaison committee with Burk, his defense team, and Ligonier leadership. Regular meetings have been occurring on Wednesday morning.

## Minister Resolution

- **redLetter:** This is a follow-up to the motion passed at the July stated meeting to immediately recognize Andrew’s membership at Saint Andrew’s Chapel pending a successful congregation vote and with withdrawal from the North Texas Presbytery.

> the next three key-values should be ignored by the scanner since there is no value provided for the key

- **motion:**
- **comments:**
- **actionItem:**

## Bank Authorization

- **motion:** (DeLoach) and seconded that the Session authorize the following changes to the list of authorized signers on the Church's bank accounts, and direct the Director of Accounting to complete the necessary documentation with the bank: **REMOVE** Stephen Adams and Lee Webb, and **ADD** Rob Bisbing, Dave Murray, Ken Moody, Bill Reisenweaver, and Andrew Sarnicki

- **actionItem:** DeLoach to collect signatures on the provided bank paperwork from the new elders and return to Stassia.

- **motion:** (Micheals) and seconded to adjourn the meeting.

- **comments:** Bill Reisenweaver closed the meeting in prayer.

> The meeting concluded at 7:30 PM

# Reminders

## September 1, 2026, Called Meeting

- **redLetter:** This meeting was called for the purpose of meeting with Burk during July's stated meeting