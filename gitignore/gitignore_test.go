package gitignore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		line     string
		baseDir  string
		expected *pattern
	}{
		// Empty and comments
		{"", "", nil},
		{"# comment", "", nil},
		{"  ", "", nil},

		// Simple patterns
		{"*.log", "", &pattern{pattern: "*.log"}},
		{"build", "", &pattern{pattern: "build"}},

		// Negation
		{"!important.log", "", &pattern{pattern: "important.log", negation: true}},

		// Directory-only
		{"build/", "", &pattern{pattern: "build", dirOnly: true}},

		// Anchored patterns
		{"/root.txt", "", &pattern{pattern: "root.txt", anchored: true}},
		{"src/main.go", "", &pattern{pattern: "src/main.go", anchored: true}},

		// With baseDir
		{"*.pyc", "subdir", &pattern{pattern: "*.pyc", baseDir: "subdir"}},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			result := parseLine(tt.line, tt.baseDir)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
				return
			}
			if result == nil {
				t.Errorf("expected %+v, got nil", tt.expected)
				return
			}
			if result.pattern != tt.expected.pattern {
				t.Errorf("pattern: expected %q, got %q", tt.expected.pattern, result.pattern)
			}
			if result.negation != tt.expected.negation {
				t.Errorf("negation: expected %v, got %v", tt.expected.negation, result.negation)
			}
			if result.dirOnly != tt.expected.dirOnly {
				t.Errorf("dirOnly: expected %v, got %v", tt.expected.dirOnly, result.dirOnly)
			}
			if result.anchored != tt.expected.anchored {
				t.Errorf("anchored: expected %v, got %v", tt.expected.anchored, result.anchored)
			}
			if result.baseDir != tt.expected.baseDir {
				t.Errorf("baseDir: expected %q, got %q", tt.expected.baseDir, result.baseDir)
			}
		})
	}
}

func TestMatchSimpleGlob(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
		match   bool
	}{
		// Exact match
		{"foo", "foo", true},
		{"foo", "bar", false},

		// Single asterisk
		{"*.go", "main.go", true},
		{"*.go", "main.py", false},
		{"test*", "test_file.go", true},
		{"*test*", "my_test_file", true},

		// Question mark
		{"?.go", "a.go", true},
		{"?.go", "ab.go", false},
		{"test?.go", "test1.go", true},

		// Asterisk doesn't match /
		{"*.go", "src/main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.name, func(t *testing.T) {
			result := matchSimpleGlob(tt.pattern, tt.name)
			if result != tt.match {
				t.Errorf("matchSimpleGlob(%q, %q) = %v, want %v", tt.pattern, tt.name, result, tt.match)
			}
		})
	}
}

func TestMatchDoublestar(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
		match   bool
	}{
		// Basic **
		{"**/*.go", "main.go", true},
		{"**/*.go", "src/main.go", true},
		{"**/*.go", "src/pkg/main.go", true},
		{"**/*.go", "main.py", false},

		// ** at end
		{"src/**", "src/file.go", true},
		{"src/**", "src/pkg/file.go", true},

		// ** in middle
		{"src/**/test.go", "src/test.go", true},
		{"src/**/test.go", "src/pkg/test.go", true},
		{"src/**/test.go", "src/a/b/test.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.name, func(t *testing.T) {
			result := matchDoublestar(tt.pattern, tt.name)
			if result != tt.match {
				t.Errorf("matchDoublestar(%q, %q) = %v, want %v", tt.pattern, tt.name, result, tt.match)
			}
		})
	}
}

func TestMatcher_Match(t *testing.T) {
	// Create a temporary directory with a .gitignore file
	tmpDir := t.TempDir()

	gitignoreContent := `
# Build outputs
build/
dist/

# Log files
*.log

# But keep important logs
!important.log

# Specific file
/root-only.txt

# Pattern with path
src/generated/
`
	err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatalf("failed to create .gitignore: %v", err)
	}

	// Create subdirectory with its own .gitignore
	subDir := filepath.Join(tmpDir, "pkg")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	subGitignore := `
# Package-specific ignores
*.tmp
`
	err = os.WriteFile(filepath.Join(subDir, ".gitignore"), []byte(subGitignore), 0644)
	if err != nil {
		t.Fatalf("failed to create sub .gitignore: %v", err)
	}

	m, err := New(tmpDir)
	if err != nil {
		t.Fatalf("failed to create matcher: %v", err)
	}

	tests := []struct {
		path   string
		isDir  bool
		ignore bool
	}{
		// Directory patterns
		{"build", true, true},
		{"build/output.exe", false, true},
		{"dist", true, true},
		{"src/dist", true, true}, // dist matches anywhere

		// File patterns
		{"debug.log", false, true},
		{"src/debug.log", false, true},
		{"important.log", false, false}, // negated

		// Anchored pattern
		{"root-only.txt", false, true},
		{"src/root-only.txt", false, false}, // doesn't match in subdirs

		// Path pattern
		{"src/generated", true, true},
		{"other/generated", true, false},

		// Subdirectory .gitignore
		{"pkg/cache.tmp", false, true},
		{"cache.tmp", false, false}, // .tmp only applies in pkg/

		// Non-ignored files
		{"main.go", false, false},
		{"src/main.go", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := m.Match(tt.path, tt.isDir)
			if result != tt.ignore {
				t.Errorf("Match(%q, %v) = %v, want %v", tt.path, tt.isDir, result, tt.ignore)
			}
		})
	}
}

func TestMatcher_Nil(t *testing.T) {
	var m *Matcher
	if m.Match("anything", false) {
		t.Error("nil matcher should not match anything")
	}
}

func TestNew_NoGitignore(t *testing.T) {
	tmpDir := t.TempDir()

	m, err := New(tmpDir)
	if err != nil {
		t.Fatalf("failed to create matcher: %v", err)
	}

	// Should not match anything without .gitignore
	if m.Match("file.txt", false) {
		t.Error("matcher without .gitignore should not match anything")
	}
}

func TestMatcher_DirectoryOnlyPattern(t *testing.T) {
	tmpDir := t.TempDir()

	gitignoreContent := `
# Only match directories named 'cache'
cache/
`
	err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatalf("failed to create .gitignore: %v", err)
	}

	m, err := New(tmpDir)
	if err != nil {
		t.Fatalf("failed to create matcher: %v", err)
	}

	// Directory should match
	if !m.Match("cache", true) {
		t.Error("directory 'cache' should be ignored")
	}

	// File should not match
	if m.Match("cache", false) {
		t.Error("file 'cache' should not be ignored (pattern has trailing /)")
	}
}
