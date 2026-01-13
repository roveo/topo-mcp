package languages

import sitter "github.com/smacker/go-tree-sitter"

// Position represents a position in a text document (LSP-compliant, 0-based)
type Position struct {
	Line      int // 0-based line number
	Character int // 0-based character offset
}

// Range represents a range in a text document (LSP-compliant, 0-based)
type Range struct {
	Start Position
	End   Position
}

// Symbol represents any code symbol across all languages
type Symbol interface {
	// Name returns the symbol's identifier
	Name() string

	// Kind returns the symbol type (e.g., "func", "method", "type", "class", "const", "var")
	Kind() string

	// Location returns the 0-based line/column range of the symbol
	Location() Range

	// String returns a language-specific rendering for the codemap
	String() string
}

// Documented is an optional interface for symbols that have documentation
type Documented interface {
	DocComment() string
}

// Language defines how to parse a particular programming language
type Language interface {
	// Name returns the language identifier (e.g., "go", "python")
	Name() string

	// Extensions returns the file extensions this language handles (e.g., [".go"])
	Extensions() []string

	// Parse parses the source content and returns imports and symbols
	Parse(content []byte) (imports []string, symbols []Symbol, err error)
}

// TreeSitterLanguage is an optional interface for languages that use tree-sitter
type TreeSitterLanguage interface {
	Language
	// TreeSitterLang returns the tree-sitter language for parsing
	TreeSitterLang() *sitter.Language
}
