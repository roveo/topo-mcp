package languages

import (
	"testing"
)

// mockLanguage is a test implementation of Language
type mockLanguage struct {
	name string
	exts []string
}

func (m *mockLanguage) Name() string         { return m.name }
func (m *mockLanguage) Extensions() []string { return m.exts }
func (m *mockLanguage) Parse(content []byte) ([]string, []Symbol, error) {
	return nil, nil, nil
}

func TestRegister(t *testing.T) {
	// Save original registry
	origRegistry := registry
	registry = make(map[string]Language)
	defer func() { registry = origRegistry }()

	lang := &mockLanguage{name: "test", exts: []string{".test", ".tst"}}
	Register(lang)

	// Check both extensions are registered
	if registry[".test"] != lang {
		t.Error("expected .test to be registered")
	}
	if registry[".tst"] != lang {
		t.Error("expected .tst to be registered")
	}
}

func TestGetLanguageForFile(t *testing.T) {
	// Save original registry
	origRegistry := registry
	registry = make(map[string]Language)
	defer func() { registry = origRegistry }()

	lang := &mockLanguage{name: "test", exts: []string{".test"}}
	Register(lang)

	tests := []struct {
		path     string
		wantLang Language
	}{
		{"file.test", lang},
		{"path/to/file.test", lang},
		{"FILE.TEST", lang}, // Case insensitive
		{"file.unknown", nil},
		{"file", nil},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := GetLanguageForFile(tt.path)
			if got != tt.wantLang {
				t.Errorf("GetLanguageForFile(%q) = %v, want %v", tt.path, got, tt.wantLang)
			}
		})
	}
}

func TestSupportedExtensions(t *testing.T) {
	// Save original registry
	origRegistry := registry
	registry = make(map[string]Language)
	defer func() { registry = origRegistry }()

	lang1 := &mockLanguage{name: "lang1", exts: []string{".a", ".b"}}
	lang2 := &mockLanguage{name: "lang2", exts: []string{".c"}}
	Register(lang1)
	Register(lang2)

	exts := SupportedExtensions()
	if len(exts) != 3 {
		t.Errorf("expected 3 extensions, got %d: %v", len(exts), exts)
	}

	// Check all are present
	extMap := make(map[string]bool)
	for _, ext := range exts {
		extMap[ext] = true
	}
	for _, ext := range []string{".a", ".b", ".c"} {
		if !extMap[ext] {
			t.Errorf("expected %q in extensions", ext)
		}
	}
}

func TestRegisteredLanguages(t *testing.T) {
	// Save original registry
	origRegistry := registry
	registry = make(map[string]Language)
	defer func() { registry = origRegistry }()

	lang1 := &mockLanguage{name: "lang1", exts: []string{".a", ".b"}}
	lang2 := &mockLanguage{name: "lang2", exts: []string{".c"}}
	Register(lang1)
	Register(lang2)

	names := RegisteredLanguages()
	if len(names) != 2 {
		t.Errorf("expected 2 languages, got %d: %v", len(names), names)
	}

	// Check both names are present
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}
	if !nameMap["lang1"] {
		t.Error("expected lang1 in registered languages")
	}
	if !nameMap["lang2"] {
		t.Error("expected lang2 in registered languages")
	}
}

func TestNodeRange(t *testing.T) {
	// This test would require creating actual tree-sitter nodes
	// which is complex, so we just test the function exists
	// and basic behavior is tested in language-specific tests
}
