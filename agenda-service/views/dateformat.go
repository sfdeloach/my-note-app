package views

import (
	"fmt"
	"time"
)

const displayDateLayout = "January 2, 2006"

// FormatDate parses an ISO 8601 (YYYY-MM-DD) date string as UTC — so a
// timezone can never shift the day — and formats it for display.
func FormatDate(iso string) (string, error) {
	t, err := time.ParseInLocation("2006-01-02", iso, time.UTC)
	if err != nil {
		return "", fmt.Errorf("views: parsing date %q: %w", iso, err)
	}
	return t.Format(displayDateLayout), nil
}
