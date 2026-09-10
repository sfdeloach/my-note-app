> Blockquotes are used like comments in code. The scanner will ignore them. For the purpose of this example meeting note, they will be used to explain the convention used to take notes. In practice, anytime I want to add text to any part of my notes that I want the scanner to ignore, I will place the text in a blockquote.
> The note's Joplin title follows the pattern `YYYY-MM-DD <Type> Meeting` — this note is titled "2026-08-11 Stated Meeting". The service cross-checks that title against the Metadata section below and errors if they disagree.
> All notes shall be divided into a `# Metadata` section plus five body sections, each an h1 header.
> key-value convention: `- **key:** value`
> regex for key-values: `^- \*\*([A-Za-z]+):\*\* (\S.*)$`
> Values are single-line. A key line with no value is skipped with a warning.
> The `# Metadata` section holds exactly one h2 — its title is ignored — containing six key-value pairs: date (ISO 8601 YYYY-MM-DD format), time, location, type, clerk, and moderator (all strings). This section is never rendered as a body section; its values populate the head matter of each view. clerk and moderator are free text, used only by the Minutes view's attestation footer.

# Metadata

## Info

- **date:** 2026-08-11
- **time:** 5:00 PM
- **location:** Classroom 7/8
- **type:** Stated
- **clerk:** RE Kevin Kennedy
- **moderator:** TE Kevin Struyk

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

# Member Updates

> The `member-updates` fenced block records the membership changes ratified at the meeting. Stanzas are separated by a blank line: line 1 is the table title, line 2 the pipe-delimited header row, the rest are pipe-delimited data rows. Valid titles are New Members, Baptisms, Transfers, Removals, and Deaths; they render in that order regardless of authoring order. Only the Minutes view renders this section — see `09-member-updates.md`.

```member-updates
New Members
First Name | Middle | Last Name | Date | Received By
Steve | J | Anyone | July 26, 2026 | Profession of Faith
Sally | L | Anyone | July 26, 2026 | Profession of Faith
Sammy | A | Anyone | July 26, 2026 | Profession of Faith
Serge | O | Anyone | July 26, 2026 | Profession of Faith
Sarah | T | Anyone | July 26, 2026 | Profession of Faith

Baptisms
First Name | Middle | Last Name | Date Baptized | Baptism Type | Parents
Jonathan | Ransom | Peeler | August 16, 2026 | Non-communing | Daniel & Bonnie Peeler

Transfers
First Name | Middle | Last Name | Transfer Date | Transfer To
Peter | J | Benyola | August 1, 2026 | St. Paul's PCA

Removals
First Name | Middle | Last Name | Removed | Reason
Sarah | F | Fowler | July 12, 2026 | non-attendance > 1 year

Deaths
First Name | Middle | Last Name | Date
Bob | K | Moser | July 21, 2026
```