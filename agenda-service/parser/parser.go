// Package parser turns a Joplin note body (Markdown, following the Agenda
// Service authoring convention) into a tree of Blocks. It knows nothing
// about Joplin, Postgres, or settings — it only understands the h1 -> h2 ->
// key-value structure and the blockquote-as-comment convention.
package parser

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

type lineType struct {
	key   string
	depth int
	re    *regexp.Regexp
}

var (
	headingRule = lineType{"h1", 0, regexp.MustCompile(`^# (.+)$`)}
	itemRule    = lineType{"item", 1, regexp.MustCompile(`^## (.+)$`)}
	keywordRule = lineType{"", 2, regexp.MustCompile(`^- \*\*([A-Za-z]+):\*\* (\S.*)$`)}

	rules = []lineType{headingRule, itemRule, keywordRule}

	valuelessKeyRule = regexp.MustCompile(`^- \*\*([A-Za-z]+):\*\*\s*$`)
	safeSkipRegex    = regexp.MustCompile(`^>.*$|^\s*$`)
)

// Block is a node in the parsed note tree: an h1 section, an h2 item, or a
// key-value leaf.
type Block struct {
	Key      string   `json:"key"`
	Content  string   `json:"content"`
	Children []*Block `json:"children"`
}

// Warning is a non-fatal condition found while parsing, such as a
// key-value line with no value.
type Warning struct {
	Line    int
	Key     string
	Message string
}

// Parse scans a note body into a tree of Blocks plus any warnings. It never
// panics or exits the process; a line that can't be classified is reported
// as an error naming the line number.
func Parse(body string) ([]*Block, []Warning, error) {
	var (
		blocks   []*Block
		warnings []Warning
		stack    []*Block // stack[i] = current open ancestor at depth i
	)

	scanner := bufio.NewScanner(strings.NewReader(body))
	lineNumber := 1

	for scanner.Scan() {
		line := scanner.Text()

		matched := false
		for _, rule := range rules {
			match := rule.re.FindStringSubmatch(line)
			if match == nil {
				continue
			}
			matched = true

			var key, content string
			if rule.key != "" {
				key, content = rule.key, match[1]
			} else {
				key, content = match[1], strings.TrimSpace(match[2])
			}

			block := &Block{Key: key, Content: content}

			if rule.depth > len(stack) {
				return nil, nil, fmt.Errorf("line %d skips a nesting level: %s", lineNumber, line)
			}
			stack = stack[:rule.depth]

			if len(stack) == 0 {
				blocks = append(blocks, block)
			} else {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, block)
			}

			stack = append(stack, block)
			break
		}

		if !matched {
			switch {
			case valuelessKeyRule.MatchString(line):
				key := valuelessKeyRule.FindStringSubmatch(line)[1]
				warnings = append(warnings, Warning{
					Line:    lineNumber,
					Key:     key,
					Message: fmt.Sprintf("line %d: key %q has no value, skipped", lineNumber, key),
				})
			case safeSkipRegex.MatchString(line):
			default:
				return nil, nil, fmt.Errorf("unrecognized line at %d: %s", lineNumber, line)
			}
		}

		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("error scanning note body: %w", err)
	}

	return blocks, warnings, nil
}
