```html
<!-- Look at the comments for notes about the new CSS and table feature -->
<!doctype html>
<html lang="en">

<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Session Meeting Minutes</title>
    <style>
        @page {
            size: letter;

            margin: 0.4in;
        }


        html {
            background: #e8e8e8;
        }

        body {
            font-family: Cambria, Cochin, Georgia, Times, "Times New Roman", serif;
            font-size: 11pt;
            line-height: 1.35;
            color: #000;
            background: #fff;
            width: 7.7in;
            min-height: 10.2in;
            margin: 0.4in auto;
            box-shadow: 0 0 6px rgba(0, 0, 0, 0.25);
        }

        @media print {
            html {
                background: none;
            }


            body {
                width: auto;
                min-height: 0;
                margin: 0;
                box-shadow: none;
            }

            /* new CSS rule: ensure in the event of long tables that
            overflow to a new printed page that the table's head is
            printed at the top  */
            thead {
                display: table-header-group;
            }
        }


        /* new CSS rules for tables follow below */
        header,
        footer {
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

        table {
            border-collapse: collapse;
            margin-bottom: 1rem;
        }

        table,
        th,
        td {
            border: 1px solid #ddd;
        }

        th {
            text-align: left;
            padding: 0.5rem;
        }

        td {
            padding: 0 0.5rem;
        }
    </style>
</head>

<body>
    <header>
        <div>
            <h2>Saint Andrew’s Chapel</h2>
        </div>
        <div>
            <h2>Session Meeting Minutes</h2>
            <h2>August 11, 2026</h2>
        </div>
    </header>

    <p>
        The Saint Andrew&rsquo;s Session held a
        <strong>stated meeting</strong> in Classroom 7/8. The meeting
        was called to order at 5:00 PM.
    </p>

    <p>
        <strong> Session Members Present: </strong>
        Alan Bird, Robert Bisbing, John Clendinen, Michael Crotty, Steve DeLoach Jr., John Enslow, Tom Gibbons, Kevin
        Kennedy, Don McDade, Chuck Micheals, Ken Moody, Dave Murray, Bill Reisenweaver, Andrew Sarnicki, Kevin Struyk,
        Lee Webb
    </p>

    <p>
        <strong> Session Members Absent: </strong> Don Bailey, Burk Parsons
    </p>


    <p>The Session spent time in corporate prayer before the meeting began.</p>

    <p><strong>Motion</strong> (Crotty) seconded to appoint Kevin Struyk to serve as the moderator in Dr.
        Parsons’absence, carried. The moderator declared that a <strong>quorum was present</strong> and called the
        meeting to order at 5:00 p.m.</p>

    <p>The agenda was presented for adoption. Bird requested adding an item to New Business: “Discussion on Status Quo.”
        The moderator requested that an item be added to New Business, “Elder Group Assignments.”</p>

    <p><strong>Motion</strong> by unanimous consent to approve the Session meeting minutes from July 21 (last stated
        meeting), August 4 (email to approve Jim Fitzgerald to preach an installation service), in omnibus, carried.</p>

    <p><strong>Motion</strong> by unanimous consent to approve the July 26 Congregational Meeting Minutes, carried.</p>

    <p>The May 2026 Deacon meeting minutes were examined.</p>

    <p><strong>Attendance and giving updates</strong> were reviewed and accepted as presented.</p>

    <p><strong>Motion</strong> (Micheals) and seconded to accept the membership updates (see ‘Member Updates’ below),
        carried.</p>

    <p><strong>Observance of the Lord’s Supper:</strong> Communion was served during the evening service on July 19,
        2026, and during the morning services on August 2, 2026.</p>

    <p><strong>Installation of elders:</strong> Ruling elders Robert Bisbing, Ken Moody, Dave Murray, and Bill
        Reisenweaver were installed during the second morning worship service on August 9, 2026. Teaching elder Andrew
        Sarnicki was installed during the evening worship service on August 9, 2026.</p>

    <p><strong>Motion</strong> (DeLoach) to form a committee of elders to liaise with Burk’s defense team and personnel
        from Ligonier, carried. The committee members include Kevin Struyk, Kevin Kennedy, and Steve DeLoach.</p>

    <p>Kevin Kennedy, on behalf of the committee, presented to the Session the recommendation to not pursue ECFA at this
        time.</p>

    <p><strong>Motion</strong> (Committee) after exploring the costs and benefits associated with association with the
        ECFA, that the Session not pursue membership at this time, carried.</p>

    <p><strong>Motion</strong> (Committee) that the Session approve the Budget Committee’s recommended reductions
        totaling $280,850, and grant staff discretionary authority to exceed budget by up to $10,000 in the aggregate
        for unforeseen expenses, carried.</p>

    <p>The Session received an update regarding a member of the church and agreed that a pastoral warning would be
        appropriate. No formal motion was entered on the matter.</p>

    <p><strong>Motion</strong> (Ad Hoc Youth &amp; Family Committee) that the Session extend Rob Conter’s contract to
        lead Wednesday night studies for one additional month, through September 30, 2026, under the same terms as the
        original agreement, to allow the Family &amp; Youth Committee time to evaluate long-term staffing options in
        light of the incoming Associate Pastor’s arrival and current budget considerations, with a recommendation to be
        brought back to Session before the extension ends, carried.</p>

    <p><strong>Motion</strong> (Gibbons) and seconded to approve the proposed preaching schedule for September and
        October 2026, carried.</p>

    <p><strong>Motion</strong> (Micheals) and seconded that a bylaw revision committee be formed to review and recommend
        changes to the existing bylaws in light of the church’s current circumstances, including recommendations
        regarding conflicts of interest, with Elder Micheals to serve as chairman.</p>

    <p>Discussion followed, during which the Session was reminded that a bylaw revision committee had previously been
        formed in 2023, when the church entered the PCA, and that this committee remained in existence. Two of the
        original members were no longer serving on the Session; the remaining original members were Kevin Struyk
        (moderator), Steve DeLoach (secretary), and Mike Crotty.</p>

    <p><strong>Motion</strong> (Kennedy) and seconded to amend the main motion by striking the proposal to form a new
        committee and providing instead that the existing 2023 bylaw revision committee be retained in its current
        composition (Kevin Struyk, moderator; Steve DeLoach, secretary; Mike Crotty), with Chuck Micheals and Ken Moody
        added as members, this reconstituted committee to take up the concerns raised by Elder Micheals. The amendment
        was adopted.</p>

    <p>The motion as amended — that the existing bylaw revision committee, consisting of Kevin Struyk (moderator), Steve
        DeLoach (secretary), and Mike Crotty, be retained and that Chuck Micheals and Ken Moody be added as members,
        with the committee to address the concerns and proposed changes raised by Elder Micheals, including
        recommendations on conflicts of interest — was adopted.</p>

    <p>The Session received and discussed the substance of a letter signed by four member families.</p>

    <p><strong>Motion</strong> (DeLoach) and seconded to form a committee of elders to meet with the members who wrote
        the letter, carried. The committee includes Kevin Kennedy, John Clendinen, Don Bailey, and Ken Moody.</p>

    <p><strong>Motion</strong> (DeLoach) and seconded to adopt the resolution recording the reception and installation
        of Andrew Sarnicki as a minister of the Word and sacraments, carried unanimously with no abstentions. See below
        for the “Resolution of the Session.”</p>

    <p><strong>Motion</strong> (DeLoach) and seconded to approve the housing allowance for Pastor Andrew Sarnicki for
        the amount of $27,000 for the remainder of 2026, carried unanimously with no abstentions.</p>

    <p><strong>Motion</strong> (Kennedy) and seconded that the Session authorize the following changes to the list of
        authorized signers on the Church’s bank accounts, and direct the Director of Accounting to complete the
        necessary documentation with the bank: <strong>REMOVE</strong> Stephen Adams and Lee Webb, and
        <strong>ADD</strong> Rob Bisbing, Dave Murray, Ken Moody, Bill Reisenweaver, and Andrew Sarnicki
    </p>

    <p>The Session discussed the need to distribute members to the new elders. Existing ruling elders were asked to
        provide six to ten families to Don McDade for distribution to the new ruling elders.</p>

    <p><strong>Motion</strong> (Micheals) and seconded to adjourn the meeting.</p>

    <p>Andrew Sarnicki closed the meeting in prayer.</p>

    <!-- the new feature's example begins here -->
    <h2>Member Updates</h2>

    <h3>New Members</h3>

    <table>
        <thead>
            <tr>
                <th>First Name</th>
                <th>Middle</th>
                <th>Last Name</th>
                <th>Date</th>
                <th>Received By</th>
            </tr>
        </thead>
        <tbody>
            <tr>
                <td>Steve</td>
                <td>J</td>
                <td>Anyone</td>
                <td>July 26, 2026</td>
                <td>Profession of Faith</td>
            </tr>
            <tr>
                <td>Sally</td>
                <td>L</td>
                <td>Anyone</td>
                <td>July 26, 2026</td>
                <td>Profession of Faith</td>
            </tr>
            <tr>
                <td>Sammy</td>
                <td>A</td>
                <td>Anyone</td>
                <td>July 26, 2026</td>
                <td>Profession of Faith</td>
            </tr>
            <tr>
                <td>Serge</td>
                <td>O</td>
                <td>Anyone</td>
                <td>July 26, 2026</td>
                <td>Profession of Faith</td>
            </tr>
            <tr>
                <td>Sarah</td>
                <td>T</td>
                <td>Anyone</td>
                <td>July 26, 2026</td>
                <td>Profession of Faith</td>
            </tr>
        </tbody>
    </table>

    <h3>Baptisms</h3>

    <table>
        <thead>
            <tr>
                <th>First Name</th>
                <th>Middle</th>
                <th>Last Name</th>
                <th>Date Baptized</th>
                <th>Baptism Type</th>
                <th>Parents</th>
            </tr>
        </thead>
        <tbody>
            <tr>
                <td>Jonathan</td>
                <td>Ransom</td>
                <td>Peeler</td>
                <td>August 16, 2026</td>
                <td>Non-communing</td>
                <td>Daniel & Bonnie Peeler</td>
            </tr>
        </tbody>
    </table>

    <h3>Transfers</h3>

    <table>
        <thead>
            <tr>
                <th>First Name</th>
                <th>Middle</th>
                <th>Last Name</th>
                <th>Transfer Date</th>
                <th>Transfer To</th>
            </tr>
        </thead>
        <tbody>
            <tr>
                <td>Peter</td>
                <td>J</td>
                <td>Benyola</td>
                <td>August 1, 2026</td>
                <td>St. Paul's PCA</td>
            </tr>
        </tbody>
    </table>

    <h3>Removals</h3>

    <table>
        <thead>
            <tr>
                <th>First Name</th>
                <th>Middle</th>
                <th>Last Name</th>
                <th>Removed</th>
                <th>Reason</th>
            </tr>
        </thead>
        <tbody>
            <tr>
                <td>Sarah</td>
                <td>F</td>
                <td>Fowler</td>
                <td>July 12, 2026</td>
                <td>non-attendance > 1 year</td>
            </tr>
        </tbody>
    </table>

    <h3>Deaths</h3>

    <table>
        <thead>
            <tr>
                <th>First Name</th>
                <th>Middle</th>
                <th>Last Name</th>
                <th>Date</th>
            </tr>
        </thead>
        <tbody>
            <tr>
                <td>Bob</td>
                <td>K</td>
                <td>Moser</td>
                <td>July 21, 2026</td>
            </tr>
        </tbody>
    </table>
    <!-- the new feature's example ends here -->

    <footer>
        <div>
            <p>Attested by Clerk RE Kevin Kennedy</p>
            <p>Signature: ___________________________</p>
            <p>Date: _______________</p>
        </div>
        <div>
            <p>Attested by Moderator TE Kevin Struyk</p>
            <p>Signature: ___________________________</p>
            <p>Date: _______________</p>
        </div>
    </footer>
</body>

</html>
```