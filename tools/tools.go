// Package tools provides MCP tool implementations for code intelligence.
package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/roveo/topo-mcp/gitignore"
	"github.com/roveo/topo-mcp/languages"
)

// DefaultLineLimit is the default maximum number of lines in the codemap output
const DefaultLineLimit = 1000

// Config holds server-wide configuration for tools
type Config struct {
	SkipPatterns []string // Path prefixes to skip by default
	LineLimit    int      // Maximum lines in output (0 = no limit)
}

// FileIndex represents the index of a single source file
type FileIndex struct {
	Path     string             `json:"path"`              // Relative path from index root
	Language string             `json:"language"`          // Language identifier (e.g., "go", "python")
	Imports  []string           `json:"imports,omitempty"` // Import paths/modules
	Symbols  []languages.Symbol `json:"-"`                 // Symbols in the file
}

// IndexDirectory walks the directory and indexes all supported source files
// IndexDirectory walks the directory and indexes all supported source files
func IndexDirectory(dir string) ([]FileIndex, error) {
	var results []FileIndex

	// Load gitignore patterns
	gitignoreMatcher, _ := gitignore.New(dir)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path for gitignore matching
		relPath, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			relPath = path
		}

		// Skip hidden directories and vendor
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			// Check gitignore for directories
			if gitignoreMatcher != nil && gitignoreMatcher.Match(relPath, true) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check gitignore for files
		if gitignoreMatcher != nil && gitignoreMatcher.Match(relPath, false) {
			return nil
		}

		// Get the language for this file
		lang := languages.GetLanguageForFile(path)
		if lang == nil {
			// Unsupported file type, skip
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			// Skip files that can't be read
			return nil
		}

		// Parse the file
		imports, symbols, err := lang.Parse(content)
		if err != nil {
			// Skip files that can't be parsed
			return nil
		}

		results = append(results, FileIndex{
			Path:     relPath,
			Language: lang.Name(),
			Imports:  imports,
			Symbols:  symbols,
		})

		return nil
	})

	return results, err
}

// ParseFile parses a single file and returns its symbols
func ParseFile(filePath string) ([]languages.Symbol, error) {
	lang := languages.GetLanguageForFile(filePath)
	if lang == nil {
		return nil, fmt.Errorf("unsupported file type: %s", filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	_, symbols, err := lang.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	return symbols, nil
}

// FindSymbol finds a symbol by name in a file
// Returns the symbol and the file content lines for that symbol
func FindSymbol(filePath string, symbolName string) (languages.Symbol, []string, error) {
	symbols, err := ParseFile(filePath)
	if err != nil {
		return nil, nil, err
	}

	// Find the symbol
	var found languages.Symbol
	for _, sym := range symbols {
		if sym.Name() == symbolName {
			found = sym
			break
		}
	}

	if found == nil {
		return nil, nil, fmt.Errorf("symbol %q not found in %s", symbolName, filePath)
	}

	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Extract the lines for the symbol
	lines := strings.Split(string(content), "\n")
	loc := found.Location()
	startLine := loc.Start.Line
	endLine := loc.End.Line

	// Bounds check
	if startLine < 0 {
		startLine = 0
	}
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	return found, lines[startLine : endLine+1], nil
}
