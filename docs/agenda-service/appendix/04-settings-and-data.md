This lives at `agenda-service/config/settings.json`, bind-mounted read-only and hand-edited as required. It holds what doesn't belong in the notes because it changes infrequently.

It is validated at startup (fatal on failure) and re-read per request, so a roster edit takes effect without restarting the container.

Explanation:

- `listType` — keyed by the **exact h1 section title** as it appears in the note. A section not listed here defaults to `unordered`.
- `elderColumns` — the number of columns the elder names are organized in for the printed agenda's roll call.
- `firstName`, `lastName`, `suffix` — names render as `First Last` or `First Last Suffix` (no comma — the Minutes rosters are comma-separated inline lists, where "Steve DeLoach, Jr." would read as two people), sorted alphabetically by last name then first name.
- `activeAt` — the date the elder became active on the Session.
- `removedAt` — the elder's **last active day**. A meeting held on this exact date still includes him; the day after does not.
- `elderClass` — teaching or ruling. Captured but not currently used in any view.

Which elders appear on the printed agenda (`metadata.date` comes from the note's Metadata section):

```
// Pseudocode. removedAt is the LAST ACTIVE DAY, hence <=.
if ((removedAt == undefined) OR (metadata.date <= removedAt)) AND (activeAt <= metadata.date)
  the elder is active, and his name should appear on the printed agenda
else
  the elder is inactive, and his name should not appear on the printed agenda
```

The same active-at-date set is the roster the Minutes view uses to derive Present and Absent.

Full example of the settings and data source in `JSON`:

```json
{
    "settings": {
        "listType": {
            "Notes": "unordered",
            "Absences": "ordered",
            "Reports & Updates": "ordered",
            "New Business": "ordered",
            "Reminders": "unordered"
        },
        "elderColumns": 3
    },
    "data": {
        "elders": [
            {
                "firstName": "Burk",
                "lastName": "Parsons",
                "activeAt": "2004-07-18",
                "elderClass": "teaching"
            },
            {
                "firstName": "Kevin",
                "lastName": "Struyk",
                "activeAt": "2011-11-06",
                "elderClass": "teaching"
            },
            {
                "firstName": "Don",
                "lastName": "Bailey",
                "activeAt": "2013-11-10",
                "elderClass": "teaching"
            },
            {
                "firstName": "Andrew",
                "lastName": "Sarnicki",
                "activeAt": "2026-08-09",
                "elderClass": "teaching"
            },
            {
                "firstName": "Kevin",
                "lastName": "Kennedy",
                "activeAt": "2005-01-23",
                "elderClass": "ruling"
            },
            {
                "firstName": "John",
                "lastName": "Enslow",
                "activeAt": "2014-03-09",
                "elderClass": "ruling"
            },
            {
                "firstName": "Don",
                "lastName": "McDade",
                "activeAt": "2010-04-25",
                "elderClass": "ruling"
            },
            {
                "firstName": "Chuck",
                "lastName": "Micheals",
                "activeAt": "2011-08-14",
                "elderClass": "ruling"
            },
            {
                "firstName": "Michael",
                "lastName": "Crotty",
                "activeAt": "2014-03-09",
                "elderClass": "ruling"
            },
            {
                "firstName": "Steve",
                "lastName": "DeLoach",
                "suffix": "Jr.",
                "activeAt": "2021-10-31",
                "elderClass": "ruling"
            },
            {
                "firstName": "David",
                "lastName": "Zima",
                "activeAt": "2014-03-18",
                "removedAt": "2026-03-09",
                "elderClass": "ruling"
            },
            {
                "firstName": "Rey",
                "lastName": "Villavicencio",
                "activeAt": "2014-12-14",
                "removedAt": "2025-08-25",
                "elderClass": "ruling"
            },
            {
                "firstName": "Lee",
                "lastName": "Webb",
                "activeAt": "2023-05-28",
                "elderClass": "ruling"
            },
            {
                "firstName": "John",
                "lastName": "Clendinen",
                "activeAt": "2023-05-28",
                "elderClass": "ruling"
            },
            {
                "firstName": "Alan",
                "lastName": "Bird",
                "activeAt": "2024-03-17",
                "elderClass": "ruling"
            },
            {
                "firstName": "Tom",
                "lastName": "Gibbons",
                "activeAt": "2024-03-17",
                "elderClass": "ruling"
            },
            {
                "firstName": "Robert",
                "lastName": "Bisbing",
                "activeAt": "2026-08-09",
                "elderClass": "ruling"
            },
            {
                "firstName": "Dave",
                "lastName": "Murray",
                "activeAt": "2026-08-09",
                "elderClass": "ruling"
            },
            {
                "firstName": "Bill",
                "lastName": "Reisenweaver",
                "activeAt": "2026-08-09",
                "elderClass": "ruling"
            },
            {
                "firstName": "Ken",
                "lastName": "Moody",
                "activeAt": "2026-08-09",
                "elderClass": "ruling"
            }
        ]
    }
}
```

At the example note's date of 2026-08-11 this yields 18 active elders — Zima and Villavicencio are both past their `removedAt`.

**Absence name matching (Minutes view):** the h2 titles under `# Absences` are matched case-insensitively against `"{firstName} {lastName}"`. The suffix is **not** part of the match key, so "Steve DeLoach" matches the record rendered as "Steve DeLoach Jr." A title that matches no elder is a hard error naming the unmatched title.