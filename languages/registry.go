package languages

import (
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

var registry = make(map[string]Language)

// Register adds a language to the registry
func Register(lang Language) {
	for _, ext := range lang.Extensions() {
		registry[ext] = lang
	}
}

// GetLanguageForFile returns the Language for a file based on its extension.
// Returns nil if the file type is not supported.
func GetLanguageForFile(path string) Language {
	ext := strings.ToLower(filepath.Ext(path))
	return registry[ext]
}

// SupportedExtensions returns all registered file extensions
func SupportedExtensions() []string {
	exts := make([]string, 0, len(registry))
	for ext := range registry {
		exts = append(exts, ext)
	}
	return exts
}

// RegisteredLanguages returns the names of all registered languages
func RegisteredLanguages() []string {
	seen := make(map[string]bool)
	var names []string
	for _, lang := range registry {
		if !seen[lang.Name()] {
			seen[lang.Name()] = true
			names = append(names, lang.Name())
		}
	}
	return names
}

// GetTreeSitterLanguage returns the tree-sitter language for a registered language name.
// Returns nil if the language doesn't exist or doesn't implement TreeSitterLanguage.
func GetTreeSitterLanguage(name string) *sitter.Language {
	for _, lang := range registry {
		if lang.Name() == name {
			if tsLang, ok := lang.(TreeSitterLanguage); ok {
				return tsLang.TreeSitterLang()
			}
			return nil
		}
	}
	return nil
}
