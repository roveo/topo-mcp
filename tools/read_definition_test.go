package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	// Import Go language parser for tests
	_ "github.com/roveo/topo-mcp/languages/golang"
)

func TestFindSymbol(t *testing.T) {
	// Create a temporary Go file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main

// Hello prints a greeting
func Hello(name string) string {
	return "Hello, " + name
}

type Person struct {
	Name string
	Age  int
}

func (p Person) Greet() string {
	return "Hi, I'm " + p.Name
}
`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	tests := []struct {
		name       string
		symbolName string
		wantErr    bool
		wantLines  int
	}{
		{
			name:       "find function",
			symbolName: "Hello",
			wantErr:    false,
			wantLines:  3, // func + body + closing brace
		},
		{
			name:       "find type",
			symbolName: "Person",
			wantErr:    false,
			wantLines:  4, // type + fields + closing brace
		},
		{
			name:       "find method",
			symbolName: "Greet",
			wantErr:    false,
			wantLines:  3,
		},
		{
			name:       "symbol not found",
			symbolName: "NotExists",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sym, lines, err := FindSymbol(testFile, tt.symbolName)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if sym == nil {
				t.Errorf("expected symbol, got nil")
				return
			}
			if sym.Name() != tt.symbolName {
				t.Errorf("symbol name = %q, want %q", sym.Name(), tt.symbolName)
			}
			if len(lines) != tt.wantLines {
				t.Errorf("got %d lines, want %d\nlines: %v", len(lines), tt.wantLines, lines)
			}
		})
	}
}

func TestFindSymbol_UnsupportedFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("hello"), 0o644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, _, err = FindSymbol(testFile, "foo")
	if err == nil {
		t.Error("expected error for unsupported file type")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("error should mention unsupported, got: %v", err)
	}
}

func TestFindSymbol_FileNotFound(t *testing.T) {
	_, _, err := FindSymbol("/nonexistent/file.go", "foo")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParseFile(t *testing.T) {
	// Create a temporary Go file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main

func Foo() {}
func Bar() {}
var X = 1
`
	err := os.WriteFile(testFile, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	symbols, err := ParseFile(testFile)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}

	if len(symbols) != 3 {
		t.Errorf("expected 3 symbols, got %d", len(symbols))
	}

	// Check symbol names
	names := make(map[string]bool)
	for _, s := range symbols {
		names[s.Name()] = true
	}
	for _, want := range []string{"Foo", "Bar", "X"} {
		if !names[want] {
			t.Errorf("missing symbol %q", want)
		}
	}
}
