package tools

import (
	"os"
	"path/filepath"
	"testing"

	// Import Go language parser for tests
	_ "github.com/roveo/topo-mcp/languages/golang"
)

func TestFindReferences(t *testing.T) {
	// Create a temporary directory with Go files
	tmpDir := t.TempDir()

	// Create main.go
	mainGo := `package main

import "fmt"

func main() {
	msg := Hello("World")
	fmt.Println(msg)
	Hello("Again")
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644)
	if err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	// Create hello.go
	helloGo := `package main

// Hello returns a greeting
func Hello(name string) string {
	return "Hello, " + name
}

func Goodbye(name string) string {
	return Hello(name) + " Goodbye!"
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "hello.go"), []byte(helloGo), 0644)
	if err != nil {
		t.Fatalf("failed to write hello.go: %v", err)
	}

	// Find references to "Hello"
	refs, err := FindReferences(tmpDir, "Hello")
	if err != nil {
		t.Fatalf("FindReferences error: %v", err)
	}

	// Should find:
	// - main.go: 2 calls (Hello("World"), Hello("Again"))
	// - hello.go: 1 definition + 1 call in Goodbye
	if len(refs) < 4 {
		t.Errorf("expected at least 4 references, got %d", len(refs))
		for _, ref := range refs {
			t.Logf("  %s:%d:%d %s", ref.File, ref.Line, ref.Column, ref.Context)
		}
	}

	// Verify we found references in both files
	files := make(map[string]int)
	for _, ref := range refs {
		files[ref.File]++
	}

	if files["main.go"] < 2 {
		t.Errorf("expected at least 2 references in main.go, got %d", files["main.go"])
	}
	if files["hello.go"] < 2 {
		t.Errorf("expected at least 2 references in hello.go, got %d", files["hello.go"])
	}
}

func TestFindReferences_NoMatches(t *testing.T) {
	tmpDir := t.TempDir()

	mainGo := `package main

func main() {
	println("hello")
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	refs, err := FindReferences(tmpDir, "NotExists")
	if err != nil {
		t.Fatalf("FindReferences error: %v", err)
	}

	if len(refs) != 0 {
		t.Errorf("expected 0 references, got %d", len(refs))
	}
}

func TestFindReferences_IgnoresStrings(t *testing.T) {
	tmpDir := t.TempDir()

	mainGo := `package main

func main() {
	// This is a comment mentioning Hello
	msg := "Hello is a function"
	println(msg)
}

func Hello() {}
`
	err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	refs, err := FindReferences(tmpDir, "Hello")
	if err != nil {
		t.Fatalf("FindReferences error: %v", err)
	}

	// Should only find the function definition, not the string or comment
	if len(refs) != 1 {
		t.Errorf("expected 1 reference (definition only), got %d", len(refs))
		for _, ref := range refs {
			t.Logf("  %s:%d:%d %s", ref.File, ref.Line, ref.Column, ref.Context)
		}
	}
}

func TestFindReferences_Types(t *testing.T) {
	tmpDir := t.TempDir()

	mainGo := `package main

type Person struct {
	Name string
}

func NewPerson(name string) *Person {
	return &Person{Name: name}
}

func (p *Person) Greet() string {
	return "Hi, " + p.Name
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	refs, err := FindReferences(tmpDir, "Person")
	if err != nil {
		t.Fatalf("FindReferences error: %v", err)
	}

	// Should find:
	// - type definition
	// - return type *Person
	// - &Person{...}
	// - receiver (p *Person)
	if len(refs) < 4 {
		t.Errorf("expected at least 4 references to Person, got %d", len(refs))
		for _, ref := range refs {
			t.Logf("  %s:%d:%d %s", ref.File, ref.Line, ref.Column, ref.Context)
		}
	}
}

func TestFindReferences_Subdirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "pkg")
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Create main.go in root
	mainGo := `package main

func main() {
	Shared()
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644)
	if err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	// Create pkg/shared.go
	sharedGo := `package pkg

func Shared() {
	println("shared")
}
`
	err = os.WriteFile(filepath.Join(subDir, "shared.go"), []byte(sharedGo), 0644)
	if err != nil {
		t.Fatalf("failed to write shared.go: %v", err)
	}

	refs, err := FindReferences(tmpDir, "Shared")
	if err != nil {
		t.Fatalf("FindReferences error: %v", err)
	}

	// Should find references in both files
	if len(refs) != 2 {
		t.Errorf("expected 2 references, got %d", len(refs))
		for _, ref := range refs {
			t.Logf("  %s:%d:%d %s", ref.File, ref.Line, ref.Column, ref.Context)
		}
	}

	// Verify paths are relative
	for _, ref := range refs {
		if filepath.IsAbs(ref.File) {
			t.Errorf("expected relative path, got absolute: %s", ref.File)
		}
	}
}
