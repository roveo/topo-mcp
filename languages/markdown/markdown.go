package markdown

import (
	"strings"

	"github.com/roveo/topo-mcp/languages"
)

func init() {
	languages.Register(&Language{})
}

// Language implements the Markdown language parser
type Language struct{}

func (l *Language) Name() string {
	return "markdown"
}

func (l *Language) Extensions() []string {
	return []string{".md", ".markdown"}
}

// Parse parses markdown content and extracts headings as symbols.
// Each heading's range extends from its line to just before the next heading
// at the same or higher level (fewer #s), or to the end of the file.
func (l *Language) Parse(content []byte) ([]string, []languages.Symbol, error) {
	lines := strings.Split(string(content), "\n")

	// First pass: find all headings with their line numbers and levels
	type headingInfo struct {
		line  int
		level int
		text  string
	}
	var headings []headingInfo

	inCodeBlock := false
	for lineNum, line := range lines {
		// Track fenced code blocks (``` or ~~~)
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inCodeBlock = !inCodeBlock
			continue
		}

		// Skip lines inside code blocks
		if inCodeBlock {
			continue
		}

		level, text := parseHeadingLine(line)
		if level > 0 {
			headings = append(headings, headingInfo{
				line:  lineNum,
				level: level,
				text:  text,
			})
		}
	}

	// Second pass: calculate end lines for each heading
	// A heading's range ends when we encounter a heading at the same or higher level
	var symbols []languages.Symbol

	for i, h := range headings {
		endLine := len(lines) - 1 // Default to end of file

		// Look for the next heading at same or higher level (lower or equal number)
		for j := i + 1; j < len(headings); j++ {
			if headings[j].level <= h.level {
				// End at the line before this heading
				endLine = headings[j].line - 1
				break
			}
		}

		// Calculate end character (end of the last line in range)
		endChar := 0
		if endLine >= 0 && endLine < len(lines) {
			endChar = len(lines[endLine])
		}

		symbols = append(symbols, &Heading{
			name:  h.text,
			level: h.level,
			loc: languages.Range{
				Start: languages.Position{Line: h.line, Character: 0},
				End:   languages.Position{Line: endLine, Character: endChar},
			},
		})
	}

	return nil, symbols, nil
}

// parseHeadingLine parses a line and returns the heading level (1-6) and text.
// Returns level 0 if the line is not a heading.
func parseHeadingLine(line string) (int, string) {
	trimmed := strings.TrimLeft(line, " \t")

	// Must start with #
	if !strings.HasPrefix(trimmed, "#") {
		return 0, ""
	}

	// Count consecutive # characters
	level := 0
	for _, ch := range trimmed {
		if ch == '#' {
			level++
		} else {
			break
		}
	}

	// Max level is 6
	if level > 6 {
		return 0, ""
	}

	// Must have space after # characters (or be just #s at end of line)
	rest := trimmed[level:]
	if len(rest) > 0 && rest[0] != ' ' && rest[0] != '\t' {
		return 0, ""
	}

	// Extract the heading text
	text := strings.TrimSpace(rest)

	// Remove trailing # characters (alternative heading syntax: ## Heading ##)
	text = strings.TrimRight(text, "#")
	text = strings.TrimSpace(text)

	return level, text
}
