package settings

import "sort"

// Name formats the elder's display name as "First Last" or, if Suffix is
// set, "First Last Suffix" — never with a comma, so it reads as one
// person in the Minutes' comma-separated rosters.
func (e Elder) Name() string {
	if e.Suffix == "" {
		return e.FirstName + " " + e.LastName
	}
	return e.FirstName + " " + e.LastName + " " + e.Suffix
}

// IsActiveAt reports whether the elder was active on meetingDate:
// ActiveAt is on or before meetingDate, and either RemovedAt is nil or
// meetingDate is on or before it (RemovedAt is the elder's last active
// day, so a meeting held that day still includes him). meetingDate and
// the elder's own date fields are ISO 8601 (YYYY-MM-DD) strings compared
// with plain Go string operators — safe because Load has already
// validated every date string parses as a real calendar date, and the
// fixed-width zero-padded format guarantees lexicographic order equals
// chronological order.
func (e Elder) IsActiveAt(meetingDate string) bool {
	if e.ActiveAt > meetingDate {
		return false
	}
	if e.RemovedAt != nil && meetingDate > *e.RemovedAt {
		return false
	}
	return true
}

// SortElders sorts elders in place by last name, then first name.
func SortElders(elders []Elder) {
	sort.Slice(elders, func(i, j int) bool {
		if elders[i].LastName != elders[j].LastName {
			return elders[i].LastName < elders[j].LastName
		}
		return elders[i].FirstName < elders[j].FirstName
	})
}

// ActiveElders returns the subset of s.Elders active on meetingDate,
// sorted last name then first name.
func ActiveElders(s Settings, meetingDate string) []Elder {
	var active []Elder
	for _, e := range s.Elders {
		if e.IsActiveAt(meetingDate) {
			active = append(active, e)
		}
	}
	SortElders(active)
	return active
}
