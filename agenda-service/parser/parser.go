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

	fenceOpenRule  = regexp.MustCompile("^```(\\S*)\\s*$")
	fenceCloseRule = regexp.MustCompile("^```\\s*$")
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

		inFence    bool
		fenceInfo  string
		fenceStart int
		fenceBody  strings.Builder
	)

	// emitFence appends the captured fenced block as a top-level (depth 0)
	// node. It carries the info string as its Key and the verbatim body as
	// its Content, and nothing nests under it — the stack is cleared so a
	// following key-value line does not attach to it.
	emitFence := func() {
		stack = stack[:0]
		blocks = append(blocks, &Block{Key: fenceInfo, Content: fenceBody.String()})
	}

	scanner := bufio.NewScanner(strings.NewReader(body))
	lineNumber := 1

	for scanner.Scan() {
		line := scanner.Text()

		// Fenced-block capture. Between an opening ```<info> line and the
		// next closing ``` line, every line is taken verbatim — no rule
		// matching, no "unrecognized line" error. This is what lets a
		// ```member-updates block (and any stray fenced content in a
		// legacy note) sit in a note without failing the whole parse.
		if inFence {
			if fenceCloseRule.MatchString(line) {
				emitFence()
				inFence = false
			} else {
				if fenceBody.Len() > 0 {
					fenceBody.WriteByte('\n')
				}
				fenceBody.WriteString(line)
			}
			lineNumber++
			continue
		}
		if m := fenceOpenRule.FindStringSubmatch(line); m != nil {
			inFence = true
			fenceInfo = m[1]
			fenceStart = lineNumber
			fenceBody.Reset()
			lineNumber++
			continue
		}

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

	// An unterminated fence is a non-structural slip, not a fatal one:
	// capture what we have to end of note and warn, consistent with how a
	// valueless key line is handled.
	if inFence {
		warnings = append(warnings, Warning{
			Line:    fenceStart,
			Key:     fenceInfo,
			Message: fmt.Sprintf("line %d: fenced block %q not closed, captured to end of note", fenceStart, fenceInfo),
		})
		emitFence()
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("error scanning note body: %w", err)
	}

	return blocks, warnings, nil
}
