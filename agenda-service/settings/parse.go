package settings

import (
	"encoding/json"
	"fmt"
	"time"
)

// jsonSettings mirrors the on-disk settings.json shape exactly (the
// {"settings": {...}, "data": {"elders": [...]}} envelope from
// docs/agenda-service/appendix/4-settings-and-data.md), separately from
// the flat Settings type the rest of the package exposes.
type jsonSettings struct {
	Settings struct {
		ListType     map[string]ListType `json:"listType"`
		ElderColumns int                 `json:"elderColumns"`
	} `json:"settings"`
	Data struct {
		Elders []jsonElder `json:"elders"`
	} `json:"data"`
}

type jsonElder struct {
	FirstName  string  `json:"firstName"`
	LastName   string  `json:"lastName"`
	Suffix     string  `json:"suffix"`
	ActiveAt   string  `json:"activeAt"`
	RemovedAt  *string `json:"removedAt"`
	ElderClass string  `json:"elderClass"`
}

// parse decodes and validates raw settings.json bytes into a Settings
// value. It performs no file I/O — Load is the only caller.
func parse(data []byte) (Settings, error) {
	var js jsonSettings
	if err := json.Unmarshal(data, &js); err != nil {
		return Settings{}, fmt.Errorf("parsing JSON: %w", err)
	}

	s := Settings{
		ListType:     js.Settings.ListType,
		ElderColumns: js.Settings.ElderColumns,
		Elders:       make([]Elder, len(js.Data.Elders)),
	}
	for i, je := range js.Data.Elders {
		s.Elders[i] = Elder{
			FirstName:  je.FirstName,
			LastName:   je.LastName,
			Suffix:     je.Suffix,
			ActiveAt:   je.ActiveAt,
			RemovedAt:  je.RemovedAt,
			ElderClass: je.ElderClass,
		}
	}

	if err := validate(s); err != nil {
		return Settings{}, err
	}

	return s, nil
}

const isoDateLayout = "2006-01-02"

func validDate(s string) bool {
	_, err := time.Parse(isoDateLayout, s)
	return err == nil
}

// validate enforces the structural rules settings.json must satisfy at
// load time. Every failure here is fatal — there is no partial-success
// case for a malformed settings file.
func validate(s Settings) error {
	for title, lt := range s.ListType {
		if lt != Ordered && lt != Unordered {
			return fmt.Errorf("listType %q: invalid value %q, want %q or %q", title, lt, Ordered, Unordered)
		}
	}

	if s.ElderColumns <= 0 {
		return fmt.Errorf("elderColumns must be positive, got %d", s.ElderColumns)
	}

	for i, e := range s.Elders {
		if e.FirstName == "" {
			return fmt.Errorf("elder %d: missing firstName", i)
		}
		if e.LastName == "" {
			return fmt.Errorf("elder %d (%s): missing lastName", i, e.FirstName)
		}
		if e.ElderClass == "" {
			return fmt.Errorf("elder %d (%s %s): missing elderClass", i, e.FirstName, e.LastName)
		}
		if !validDate(e.ActiveAt) {
			return fmt.Errorf("elder %d (%s %s): activeAt %q is not a valid YYYY-MM-DD date", i, e.FirstName, e.LastName, e.ActiveAt)
		}
		if e.RemovedAt != nil && !validDate(*e.RemovedAt) {
			return fmt.Errorf("elder %d (%s %s): removedAt %q is not a valid YYYY-MM-DD date", i, e.FirstName, e.LastName, *e.RemovedAt)
		}
	}

	return nil
}
