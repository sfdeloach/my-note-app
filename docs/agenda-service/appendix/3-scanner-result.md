This is the tree that `1-example-note.md` parses to. It is the parser's **test fixture** — the implementation must reproduce it exactly.

Two things to note when comparing:

- `&` appears as `\u0026` because that is what Go's `encoding/json` emits by default.
- The three valueless key lines under "Minister Resolution" produce no blocks — they are skipped with warnings.

Full JSON example:

```json
[
  {
    "key": "h1",
    "content": "Metadata",
    "children": [
      {
        "key": "item",
        "content": "Info",
        "children": [
          {
            "key": "date",
            "content": "2026-08-11",
            "children": null
          },
          {
            "key": "time",
            "content": "5:00 PM",
            "children": null
          },
          {
            "key": "location",
            "content": "Classroom 7/8",
            "children": null
          },
          {
            "key": "type",
            "content": "Stated",
            "children": null
          }
        ]
      }
    ]
  },
  {
    "key": "h1",
    "content": "Notes",
    "children": [
      {
        "key": "item",
        "content": "Elders invited to come early to prayer at 4:30 PM",
        "children": [
          {
            "key": "redLetter",
            "content": "Lee Webb would like \"to lead the Session\" in a time of prayer",
            "children": null
          }
        ]
      }
    ]
  },
  {
    "key": "h1",
    "content": "Absences",
    "children": [
      {
        "key": "item",
        "content": "Burk Parsons",
        "children": null
      },
      {
        "key": "item",
        "content": "Dave Murray",
        "children": [
          {
            "key": "redLetter",
            "content": "Traveling, notified via email on 8/1",
            "children": null
          }
        ]
      }
    ]
  },
  {
    "key": "h1",
    "content": "Reports \u0026 Updates",
    "children": [
      {
        "key": "item",
        "content": "Opening Prayer",
        "children": [
          {
            "key": "comments",
            "content": "Lee Webb opened the meeting in prayer.",
            "children": null
          }
        ]
      },
      {
        "key": "item",
        "content": "Appoint a Moderator",
        "children": [
          {
            "key": "motion",
            "content": "(Crotty) to appoint Kevin Struyk as the moderator in Burk Parsons' absence, carried.",
            "children": null
          }
        ]
      },
      {
        "key": "item",
        "content": "Membership Updates",
        "children": null
      }
    ]
  },
  {
    "key": "h1",
    "content": "New Business",
    "children": [
      {
        "key": "item",
        "content": "Communication Liaison Committee",
        "children": [
          {
            "key": "redLetter",
            "content": "A proposal that Kennedy, Struyk, and DeLoach form a liaison committee with Burk, his defense team, and Ligonier leadership. Regular meetings have been occurring on Wednesday morning.",
            "children": null
          }
        ]
      },
      {
        "key": "item",
        "content": "Minister Resolution",
        "children": [
          {
            "key": "redLetter",
            "content": "This is a follow-up to the motion passed at the July stated meeting to immediately recognize Andrew’s membership at Saint Andrew’s Chapel pending a successful congregation vote and with withdrawal from the North Texas Presbytery.",
            "children": null
          }
        ]
      },
      {
        "key": "item",
        "content": "Bank Authorization",
        "children": [
          {
            "key": "motion",
            "content": "(DeLoach) and seconded that the Session authorize the following changes to the list of authorized signers on the Church's bank accounts, and direct the Director of Accounting to complete the necessary documentation with the bank: **REMOVE** Stephen Adams and Lee Webb, and **ADD** Rob Bisbing, Dave Murray, Ken Moody, Bill Reisenweaver, and Andrew Sarnicki",
            "children": null
          },
          {
            "key": "actionItem",
            "content": "DeLoach to collect signatures on the provided bank paperwork from the new elders and return to Stassia.",
            "children": null
          },
          {
            "key": "motion",
            "content": "(Micheals) and seconded to adjourn the meeting.",
            "children": null
          },
          {
            "key": "comments",
            "content": "Bill Reisenweaver closed the meeting in prayer.",
            "children": null
          }
        ]
      }
    ]
  },
  {
    "key": "h1",
    "content": "Reminders",
    "children": [
      {
        "key": "item",
        "content": "September 1, 2026, Called Meeting",
        "children": [
          {
            "key": "redLetter",
            "content": "This meeting was called for the purpose of meeting with Burk during July's stated meeting",
            "children": null
          }
        ]
      }
    ]
  }
]
```

Expected warnings from this note:

```
line 62: key "motion" has no value, skipped
line 63: key "comments" has no value, skipped
line 64: key "actionItem" has no value, skipped
```

(Line numbers are illustrative — assert on the key names and the count, not the exact lines, so the fixture doesn't break every time the example note is edited.)