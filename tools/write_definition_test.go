package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	// Import Go language parser for tests
	_ "github.com/roveo/topo-mcp/languages/golang"
)

func TestReplaceSymbol(t *testing.T) {
	// Create a temporary Go file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main

func Hello(name string) string {
	return "Hello, " + name
}

func Goodbye(name string) string {
	return "Goodbye, " + name
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Replace Hello function
	newCode := `func Hello(name string) string {
	return "Hi, " + name + "!"
}`
	err = ReplaceSymbol(testFile, "Hello", newCode)
	if err != nil {
		t.Fatalf("ReplaceSymbol error: %v", err)
	}

	// Read back and verify
	result, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	resultStr := string(result)

	// Should contain new code
	if !strings.Contains(resultStr, `return "Hi, " + name + "!"`) {
		t.Errorf("new code not found in result:\n%s", resultStr)
	}

	// Should NOT contain old code
	if strings.Contains(resultStr, `return "Hello, " + name`) {
		t.Errorf("old code still present in result:\n%s", resultStr)
	}

	// Should still contain Goodbye function
	if !strings.Contains(resultStr, "func Goodbye") {
		t.Errorf("Goodbye function missing from result:\n%s", resultStr)
	}

	// Should still have package declaration
	if !strings.Contains(resultStr, "package main") {
		t.Errorf("package declaration missing from result:\n%s", resultStr)
	}
}

func TestReplaceSymbol_Type(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main

type Person struct {
	Name string
}

func Hello() {}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Replace Person type
	newCode := `type Person struct {
	Name string
	Age  int
}`
	err = ReplaceSymbol(testFile, "Person", newCode)
	if err != nil {
		t.Fatalf("ReplaceSymbol error: %v", err)
	}

	// Read back and verify
	result, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	resultStr := string(result)

	// Should contain Age field
	if !strings.Contains(resultStr, "Age  int") {
		t.Errorf("new field not found in result:\n%s", resultStr)
	}

	// Should still contain Hello function
	if !strings.Contains(resultStr, "func Hello()") {
		t.Errorf("Hello function missing from result:\n%s", resultStr)
	}
}

func TestReplaceSymbol_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main

func Hello() {}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err = ReplaceSymbol(testFile, "NotExists", "func NotExists() {}")
	if err == nil {
		t.Error("expected error for non-existent symbol")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestReplaceSymbol_PreservesFileStructure(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main

import "fmt"

// First function
func First() {
	fmt.Println("first")
}

// Second function
func Second() {
	fmt.Println("second")
}

// Third function
func Third() {
	fmt.Println("third")
}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Replace middle function
	newCode := `// Second function - updated
func Second() {
	fmt.Println("SECOND!")
}`
	err = ReplaceSymbol(testFile, "Second", newCode)
	if err != nil {
		t.Fatalf("ReplaceSymbol error: %v", err)
	}

	// Read back
	result, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	resultStr := string(result)

	// Verify structure is preserved
	checks := []string{
		"package main",
		`import "fmt"`,
		"// First function",
		"func First()",
		"// Second function - updated",
		`fmt.Println("SECOND!")`,
		"// Third function",
		"func Third()",
	}

	for _, check := range checks {
		if !strings.Contains(resultStr, check) {
			t.Errorf("missing expected content %q in result:\n%s", check, resultStr)
		}
	}
}
