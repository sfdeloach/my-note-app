package settings

import (
	"encoding/json"
	"testing"
)

func TestParse_WellFormed(t *testing.T) {
	data := []byte(`{
		"settings": {
			"listType": {"Notes": "unordered", "Absences": "ordered"},
			"elderColumns": 3
		},
		"data": {
			"elders": [
				{"firstName": "Burk", "lastName": "Parsons", "activeAt": "2004-07-18", "elderClass": "teaching"}
			]
		}
	}`)

	s, err := parse(data)
	if err != nil {
		t.Fatalf("parse: unexpected error: %v", err)
	}
	if s.ElderColumns != 3 {
		t.Errorf("ElderColumns = %d, want 3", s.ElderColumns)
	}
	if got := s.ListType["Absences"]; got != Ordered {
		t.Errorf("ListType[Absences] = %q, want %q", got, Ordered)
	}
	if len(s.Elders) != 1 {
		t.Fatalf("len(Elders) = %d, want 1", len(s.Elders))
	}
	if s.Elders[0].FirstName != "Burk" {
		t.Errorf("Elders[0].FirstName = %q, want Burk", s.Elders[0].FirstName)
	}
}

func TestParse_Errors(t *testing.T) {
	base := func() map[string]any {
		return map[string]any{
			"settings": map[string]any{
				"listType":     map[string]any{"Notes": "unordered"},
				"elderColumns": 3,
			},
			"data": map[string]any{
				"elders": []any{
					map[string]any{
						"firstName":  "Burk",
						"lastName":   "Parsons",
						"activeAt":   "2004-07-18",
						"elderClass": "teaching",
					},
				},
			},
		}
	}

	cases := []struct {
		name   string
		data   []byte
		modify func(m map[string]any)
	}{
		{
			name: "malformed JSON syntax",
			data: []byte(`{"settings": {`),
		},
		{
			name: "bad listType value",
			modify: func(m map[string]any) {
				m["settings"].(map[string]any)["listType"] = map[string]any{"Notes": "sideways"}
			},
		},
		{
			name: "elder missing firstName",
			modify: func(m map[string]any) {
				elder := m["data"].(map[string]any)["elders"].([]any)[0].(map[string]any)
				delete(elder, "firstName")
			},
		},
		{
			name: "elder missing lastName",
			modify: func(m map[string]any) {
				elder := m["data"].(map[string]any)["elders"].([]any)[0].(map[string]any)
				delete(elder, "lastName")
			},
		},
		{
			name: "elder malformed activeAt",
			modify: func(m map[string]any) {
				elder := m["data"].(map[string]any)["elders"].([]any)[0].(map[string]any)
				elder["activeAt"] = "2026-13-40"
			},
		},
		{
			name: "elder malformed removedAt",
			modify: func(m map[string]any) {
				elder := m["data"].(map[string]any)["elders"].([]any)[0].(map[string]any)
				elder["removedAt"] = "not-a-date"
			},
		},
		{
			name: "elderColumns zero",
			modify: func(m map[string]any) {
				m["settings"].(map[string]any)["elderColumns"] = 0
			},
		},
		{
			name: "elderColumns negative",
			modify: func(m map[string]any) {
				m["settings"].(map[string]any)["elderColumns"] = -1
			},
		},
		{
			name: "elder empty elderClass",
			modify: func(m map[string]any) {
				elder := m["data"].(map[string]any)["elders"].([]any)[0].(map[string]any)
				elder["elderClass"] = ""
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			data := c.data
			if data == nil {
				m := base()
				c.modify(m)
				encoded, err := json.Marshal(m)
				if err != nil {
					t.Fatalf("marshalling test fixture: %v", err)
				}
				data = encoded
			}
			if _, err := parse(data); err == nil {
				t.Errorf("parse: expected error, got nil")
			}
		})
	}
}

func TestParse_UnrecognizedElderClassAllowed(t *testing.T) {
	data := []byte(`{
		"settings": {"listType": {}, "elderColumns": 3},
		"data": {
			"elders": [
				{"firstName": "A", "lastName": "B", "activeAt": "2004-07-18", "elderClass": "deacon"}
			]
		}
	}`)

	if _, err := parse(data); err != nil {
		t.Errorf("parse: unexpected error for unrecognized-but-nonempty elderClass: %v", err)
	}
}
