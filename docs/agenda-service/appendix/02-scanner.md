This is my existing prototype: a Go function that parses the Markdown note into a working tree structure. **It is the starting point, not the deliverable.**

Two things about it are prototype artifacts and change in the real service:

- It declares `package service`. The target is `package parser`.
- It takes an `*os.File` and calls `log.Fatalf`. The target takes the body string (or an `io.Reader`) and returns `([]*Block, []Warning, error)`.

What to keep: the depth/stack discipline, the skip-level guard, and the `Block{Key, Content, Children}` shape.

What to add: a rule for valueless key lines (`^- \*\*([A-Za-z]+):\*\*\s*$`), matched before the unrecognized-line branch and emitted as a warning. As written below, `- **motion:**` matches neither `keywordRule` (which requires a space and then `\S`) nor `safeSkipRegex` (blockquote or blank only), so it falls through to the error branch. Values also need trimming.

```go
package service

import (
	"bufio"
	"log"
	"os"
	"regexp"
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

	safeSkipRegex = regexp.MustCompile(`^>.*$|^\s*$`)
)

type Block struct {
	Key      string   `json:"key"`
	Content  string   `json:"content"`
	Children []*Block `json:"children"`
}

func Scanner(file *os.File) *[]*Block {
	blocks := []*Block{}
	stack := []*Block{} // stack[i] = current open ancestor at depth i

	scanner := bufio.NewScanner(file)
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
				key, content = match[1], match[2] // keyword rule: key/content come from the line itself
			}

			block := &Block{Key: key, Content: content}

			// Trim the stack back to this block's parent depth.
			if rule.depth > len(stack) {
				log.Fatalf("Line %d skips a nesting level: %s", lineNumber, line)
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
			case safeSkipRegex.MatchString(line):
			default:
				log.Fatalf("Unrecognized line at %d: %s", lineNumber, line)
			}
		}

		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error with scanner: %v", err)
	}

	return &blocks
}
```