package markdown

import (
	"testing"
)

func TestLanguageMetadata(t *testing.T) {
	lang := &Language{}

	if lang.Name() != "markdown" {
		t.Errorf("expected name 'markdown', got %q", lang.Name())
	}

	exts := lang.Extensions()
	if len(exts) != 2 {
		t.Errorf("expected 2 extensions, got %d", len(exts))
	}
	if exts[0] != ".md" || exts[1] != ".markdown" {
		t.Errorf("expected extensions [.md, .markdown], got %v", exts)
	}
}

func TestParseSingleHeading(t *testing.T) {
	src := `# Hello World

This is some content.
`
	lang := &Language{}
	imports, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(imports) != 0 {
		t.Errorf("expected no imports, got %v", imports)
	}

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	h := symbols[0]
	if h.Name() != "Hello World" {
		t.Errorf("expected name 'Hello World', got %q", h.Name())
	}
	if h.Kind() != "h1" {
		t.Errorf("expected kind 'h1', got %q", h.Kind())
	}
	if h.String() != "# Hello World" {
		t.Errorf("expected String() '# Hello World', got %q", h.String())
	}

	loc := h.Location()
	if loc.Start.Line != 0 {
		t.Errorf("expected start line 0, got %d", loc.Start.Line)
	}
	// Should extend to end of file (line 3 - empty line after trailing newline)
	if loc.End.Line != 3 {
		t.Errorf("expected end line 3, got %d", loc.End.Line)
	}
}

func TestParseNestedHeadings(t *testing.T) {
	src := `# Chapter 1

Intro text.

## Section 1.1

Section content.

## Section 1.2

More content.

# Chapter 2

Chapter 2 content.
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 4 {
		t.Fatalf("expected 4 symbols, got %d", len(symbols))
	}

	// Chapter 1 should end before Chapter 2 (line 12)
	ch1 := symbols[0]
	if ch1.Name() != "Chapter 1" {
		t.Errorf("expected 'Chapter 1', got %q", ch1.Name())
	}
	if ch1.Location().End.Line != 11 {
		t.Errorf("Chapter 1: expected end line 11, got %d", ch1.Location().End.Line)
	}

	// Section 1.1 should end before Section 1.2 (line 7)
	sec11 := symbols[1]
	if sec11.Name() != "Section 1.1" {
		t.Errorf("expected 'Section 1.1', got %q", sec11.Name())
	}
	if sec11.Kind() != "h2" {
		t.Errorf("expected kind 'h2', got %q", sec11.Kind())
	}
	if sec11.Location().End.Line != 7 {
		t.Errorf("Section 1.1: expected end line 7, got %d", sec11.Location().End.Line)
	}

	// Section 1.2 should end before Chapter 2 (line 11)
	sec12 := symbols[2]
	if sec12.Name() != "Section 1.2" {
		t.Errorf("expected 'Section 1.2', got %q", sec12.Name())
	}
	if sec12.Location().End.Line != 11 {
		t.Errorf("Section 1.2: expected end line 11, got %d", sec12.Location().End.Line)
	}

	// Chapter 2 should extend to end of file
	ch2 := symbols[3]
	if ch2.Name() != "Chapter 2" {
		t.Errorf("expected 'Chapter 2', got %q", ch2.Name())
	}
	if ch2.Location().End.Line != 15 {
		t.Errorf("Chapter 2: expected end line 15, got %d", ch2.Location().End.Line)
	}
}

func TestParseHeadingLevels(t *testing.T) {
	src := `# H1
## H2
### H3
#### H4
##### H5
###### H6
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 6 {
		t.Fatalf("expected 6 symbols, got %d", len(symbols))
	}

	for i, h := range symbols {
		expected := i + 1
		if h.Kind() != "h"+string(rune('0'+expected)) {
			t.Errorf("symbol %d: expected kind 'h%d', got %q", i, expected, h.Kind())
		}
	}
}

func TestParseHeadingRangesWithDeeperNesting(t *testing.T) {
	// H1 contains H2 and H3; H2 ends when H1 ends (before next H1)
	src := `# Title

## Subtitle

### Deeper

Some text

# Another Title
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 4 {
		t.Fatalf("expected 4 symbols, got %d", len(symbols))
	}

	// Title (H1) ends at line 7 (before "# Another Title" at line 8)
	title := symbols[0]
	if title.Location().End.Line != 7 {
		t.Errorf("Title: expected end line 7, got %d", title.Location().End.Line)
	}

	// Subtitle (H2) also ends at line 7 (before "# Another Title")
	subtitle := symbols[1]
	if subtitle.Location().End.Line != 7 {
		t.Errorf("Subtitle: expected end line 7, got %d", subtitle.Location().End.Line)
	}

	// Deeper (H3) also ends at line 7
	deeper := symbols[2]
	if deeper.Location().End.Line != 7 {
		t.Errorf("Deeper: expected end line 7, got %d", deeper.Location().End.Line)
	}

	// Another Title extends to end
	another := symbols[3]
	if another.Location().End.Line != 9 {
		t.Errorf("Another Title: expected end line 9, got %d", another.Location().End.Line)
	}
}

func TestParseTrailingHashes(t *testing.T) {
	src := `## Heading ##

Content
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	if symbols[0].Name() != "Heading" {
		t.Errorf("expected 'Heading', got %q", symbols[0].Name())
	}
}

func TestParseNotAHeading(t *testing.T) {
	src := `This is #not a heading

#Also not a heading

 # This is a heading with leading space

##NoSpaceAfterHashes
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Only " # This is a heading with leading space" should be parsed
	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	if symbols[0].Name() != "This is a heading with leading space" {
		t.Errorf("expected 'This is a heading with leading space', got %q", symbols[0].Name())
	}
}

func TestParseEmptyFile(t *testing.T) {
	src := ``
	lang := &Language{}
	imports, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(imports) != 0 {
		t.Errorf("expected no imports, got %v", imports)
	}
	if len(symbols) != 0 {
		t.Errorf("expected no symbols, got %v", symbols)
	}
}

func TestParseFileWithNoHeadings(t *testing.T) {
	src := `Just some text.

More text here.

- A list item
- Another item
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 0 {
		t.Errorf("expected no symbols, got %d", len(symbols))
	}
}

func TestParseHeadingWithSpecialCharacters(t *testing.T) {
	src := `# Hello, World! How's it going?

Content here.
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	if symbols[0].Name() != "Hello, World! How's it going?" {
		t.Errorf("expected \"Hello, World! How's it going?\", got %q", symbols[0].Name())
	}
}

func TestMoreThanSixHashes(t *testing.T) {
	src := `####### Not a valid heading

# Valid H1
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Only "# Valid H1" should be parsed
	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	if symbols[0].Name() != "Valid H1" {
		t.Errorf("expected 'Valid H1', got %q", symbols[0].Name())
	}
}

func TestIgnoreCodeBlocks(t *testing.T) {
	src := "# Real Heading\n\n```bash\n# This is a comment\n## Not a heading\n```\n\n## Another Real Heading\n"

	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(symbols))
	}

	if symbols[0].Name() != "Real Heading" {
		t.Errorf("expected 'Real Heading', got %q", symbols[0].Name())
	}
	if symbols[1].Name() != "Another Real Heading" {
		t.Errorf("expected 'Another Real Heading', got %q", symbols[1].Name())
	}
}

func TestIgnoreTildeCodeBlocks(t *testing.T) {
	src := "# Heading\n\n~~~\n# Comment in code\n~~~\n\n## Subheading\n"

	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(symbols))
	}

	if symbols[0].Name() != "Heading" {
		t.Errorf("expected 'Heading', got %q", symbols[0].Name())
	}
	if symbols[1].Name() != "Subheading" {
		t.Errorf("expected 'Subheading', got %q", symbols[1].Name())
	}
}

func TestHeadingLocation(t *testing.T) {
	src := `# First

Content

## Second

More content
`
	lang := &Language{}
	_, symbols, err := lang.Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(symbols) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(symbols))
	}

	// First heading starts at line 0
	first := symbols[0]
	if first.Location().Start.Line != 0 {
		t.Errorf("First: expected start line 0, got %d", first.Location().Start.Line)
	}
	if first.Location().Start.Character != 0 {
		t.Errorf("First: expected start char 0, got %d", first.Location().Start.Character)
	}

	// Second heading starts at line 4
	second := symbols[1]
	if second.Location().Start.Line != 4 {
		t.Errorf("Second: expected start line 4, got %d", second.Location().Start.Line)
	}
}
