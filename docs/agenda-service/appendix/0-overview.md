# Streamlined Notes

> **Non-normative.** This file is the original problem statement, kept for background. Where it disagrees with the build brief, the brief wins.

## Idea

Write one single "source of truth" for every Session meeting that contains all agenda items, pastor's notes, motions, notes, and action items in one markdown file. Use this single file to produce a:

1. Printed Agenda
2. Pastor's "Red-Letter" Agenda
3. Meeting Minutes
4. Action Items

**The Problem this Solves:** Traditionally, agenda items and notes leading to a meeting were kept on Microsoft OneNote. When the meeting date came closer, these notes would be copied into Microsoft Word which would require heavy editing to remove notes and formatting. Additionally, a second "red-letter agenda" would be created from a copy and paste from the same OneNote source, and it would undergo specific editing and formatting. As the agenda for any meeting can be quite fluid, it became challenging to remember to edit three separate documents and drift would be the best case scenario and outright omissions the worst. After a meeting is concluded, two important documents are produced: the minutes and an "action items" document to track what needed to be done after the meeting.

The idea is to migrate to a Joplin based note taking system that allows greater control and ownership of sensitive data, and to build tooling on top of it that automates and streamlines the above described process.

## How it Works

1. A single source of truth is maintained in a Joplin notebook for each meeting. A particular convention is adopted and repeated for each meeting using Markdown formatting to track all the information that is produced by meetings.
2. A Go service reads the master note directly from Joplin's Postgres database — Joplin *Server* has no Data API — and renders four views on demand:
   a. Printed Agenda
   b. Red Letter Agenda
   c. Meeting Minutes
   d. Action Items
3. The Markdown note is parsed into an in-memory tree, and each view is a renderer over that tree. Each document varies in its format and intended means of being shared with the Session and pastors.