package views

import (
	"strings"
	"testing"
)

func TestBold(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"no bold", "plain text", "plain text"},
		{"one pair", "**bold**", "<strong>bold</strong>"},
		{"pair in context", "before **bold** after", "before <strong>bold</strong> after"},
		{
			"two pairs (bank motion shape)",
			"**REMOVE** Stephen Adams and Lee Webb, and **ADD** Rob Bisbing",
			"<strong>REMOVE</strong> Stephen Adams and Lee Webb, and <strong>ADD</strong> Rob Bisbing",
		},
		{"unmatched trailing delimiter", "text **unclosed", "text **unclosed"},
		{"unmatched leading delimiter", "**unclosed text", "**unclosed text"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := string(Bold(c.in))
			if got != c.want {
				t.Errorf("Bold(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestBold_EscapesBeforeSubstituting(t *testing.T) {
	got := string(Bold(`<script>alert("x")</script> **bold**`))
	if strings.Contains(got, "<script>") {
		t.Errorf("Bold output contains an unescaped <script> tag: %q", got)
	}
	if !strings.Contains(got, "<strong>bold</strong>") {
		t.Errorf("Bold output missing the expected <strong>: %q", got)
	}
}
