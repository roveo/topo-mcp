package tools

import (
	"strings"
	"testing"

	"github.com/roveo/topo-mcp/languages"
)

// mockSymbol implements languages.Symbol for testing
type mockSymbol struct {
	symbolName string
	symbolKind string
	loc        languages.Range
}

func (s mockSymbol) Name() string              { return s.symbolName }
func (s mockSymbol) Kind() string              { return s.symbolKind }
func (s mockSymbol) String() string            { return s.symbolKind + " " + s.symbolName }
func (s mockSymbol) Location() languages.Range { return s.loc }

func makeTestFiles(count int, symbolsPerFile int) []FileIndex {
	files := make([]FileIndex, count)
	for i := 0; i < count; i++ {
		symbols := make([]languages.Symbol, symbolsPerFile)
		for j := 0; j < symbolsPerFile; j++ {
			symbols[j] = mockSymbol{
				symbolName: "symbol",
				symbolKind: "func",
				loc: languages.Range{
					Start: languages.Position{Line: j * 10},
					End:   languages.Position{Line: j*10 + 5},
				},
			}
		}
		files[i] = FileIndex{
			Path:     "file" + string(rune('a'+i)) + ".go",
			Language: "go",
			Symbols:  symbols,
		}
	}
	return files
}

func makeTestFilesInDirs(dirs []string, symbolsPerFile int) []FileIndex {
	var files []FileIndex
	for _, dir := range dirs {
		symbols := make([]languages.Symbol, symbolsPerFile)
		for j := 0; j < symbolsPerFile; j++ {
			symbols[j] = mockSymbol{
				symbolName: "symbol",
				symbolKind: "func",
				loc: languages.Range{
					Start: languages.Position{Line: j * 10},
					End:   languages.Position{Line: j*10 + 5},
				},
			}
		}
		path := dir + "/main.go"
		if dir == "" {
			path = "main.go"
		}
		files = append(files, FileIndex{
			Path:     path,
			Language: "go",
			Symbols:  symbols,
		})
	}
	return files
}

func TestFileLineCount(t *testing.T) {
	tests := []struct {
		name     string
		symbols  int
		expected int
	}{
		{"empty file", 0, 0},
		{"one symbol", 1, 3},   // header + 1 symbol + blank
		{"five symbols", 5, 7}, // header + 5 symbols + blank
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := makeTestFiles(1, tt.symbols)[0]
			got := fileLineCount(file)
			if got != tt.expected {
				t.Errorf("fileLineCount() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestBuildDirTree(t *testing.T) {
	files := makeTestFilesInDirs([]string{"", "cmd", "pkg/api", "pkg/util"}, 5)

	tree := buildDirTree(files, FormatOptions{})

	// Root should have files from "" and children "cmd", "pkg"
	if len(tree.files) != 1 {
		t.Errorf("root should have 1 file, got %d", len(tree.files))
	}
	if len(tree.children) != 2 {
		t.Errorf("root should have 2 children, got %d", len(tree.children))
	}

	// Check pkg has 2 children (api, util)
	pkg := tree.children["pkg"]
	if pkg == nil {
		t.Fatal("pkg directory not found")
	}
	if len(pkg.children) != 2 {
		t.Errorf("pkg should have 2 children, got %d", len(pkg.children))
	}
}

func TestCalculateLines(t *testing.T) {
	files := makeTestFilesInDirs([]string{"", "cmd", "pkg/api"}, 5)
	tree := buildDirTree(files, FormatOptions{})

	// Each file with 5 symbols = 7 lines (1 header + 5 symbols + 1 blank)
	// 3 files = 21 lines
	expectedTotal := 21

	if tree.lines != expectedTotal {
		t.Errorf("total lines = %d, want %d", tree.lines, expectedTotal)
	}
}

func TestPruneToLimit_NoLimit(t *testing.T) {
	files := makeTestFiles(5, 10) // 5 files, 10 symbols each

	tree := buildDirTree(files, FormatOptions{})
	prunedFiles, prunedDirs := pruneToLimit(tree, 0)

	if len(prunedFiles) != 5 {
		t.Errorf("expected 5 files, got %d", len(prunedFiles))
	}
	if len(prunedDirs) != 0 {
		t.Errorf("expected no pruned dirs, got %v", prunedDirs)
	}
}

func TestPruneToLimit_UnderLimit(t *testing.T) {
	files := makeTestFiles(2, 5) // 2 files, 5 symbols each = 14 lines

	tree := buildDirTree(files, FormatOptions{})
	prunedFiles, prunedDirs := pruneToLimit(tree, 100)

	if len(prunedFiles) != 2 {
		t.Errorf("expected 2 files, got %d", len(prunedFiles))
	}
	if len(prunedDirs) != 0 {
		t.Errorf("expected no pruned dirs, got %v", prunedDirs)
	}
}

func TestPruneToLimit_PrunesLargestFirst(t *testing.T) {
	// Create files in different directories with different sizes
	files := []FileIndex{
		{Path: "small/a.go", Language: "go", Symbols: makeSymbols(2)},
		{Path: "large/b.go", Language: "go", Symbols: makeSymbols(20)},
		{Path: "medium/c.go", Language: "go", Symbols: makeSymbols(10)},
	}

	// small = 4 lines, large = 22 lines, medium = 12 lines
	// Total = 38 lines
	// With limit 20, should prune "large" first (22 lines)
	tree := buildDirTree(files, FormatOptions{})
	prunedFiles, prunedDirs := pruneToLimit(tree, 20)

	// Should have small and medium, large should be pruned
	if len(prunedFiles) != 2 {
		t.Errorf("expected 2 files, got %d", len(prunedFiles))
	}

	// large directory should be pruned
	if len(prunedDirs) != 1 || prunedDirs[0] != "large" {
		t.Errorf("expected [large] to be pruned, got %v", prunedDirs)
	}
}

func makeSymbols(count int) []languages.Symbol {
	symbols := make([]languages.Symbol, count)
	for i := 0; i < count; i++ {
		symbols[i] = mockSymbol{
			symbolName: "symbol",
			symbolKind: "func",
			loc: languages.Range{
				Start: languages.Position{Line: i * 10},
				End:   languages.Position{Line: i*10 + 5},
			},
		}
	}
	return symbols
}

func TestFormatCodemap_WithLineLimit(t *testing.T) {
	// Create files in directories to trigger directory pruning
	// Each file with 20 symbols = 22 lines (header + 20 symbols + blank)
	files := []FileIndex{
		{Path: "dir1/a.go", Language: "go", Symbols: makeSymbols(20)},
		{Path: "dir2/b.go", Language: "go", Symbols: makeSymbols(20)},
		{Path: "dir3/c.go", Language: "go", Symbols: makeSymbols(20)},
		{Path: "dir4/d.go", Language: "go", Symbols: makeSymbols(20)},
		{Path: "dir5/e.go", Language: "go", Symbols: makeSymbols(20)},
	}
	// Total = 5 * 22 = 110 lines

	output := FormatCodemap(files, FormatOptions{
		LineLimit: 50,
	})

	// Output should be under limit (approximately)
	lines := strings.Split(output, "\n")
	// Allow some overhead for the pruning notice header
	if len(lines) > 60 {
		t.Errorf("output should be around 50 lines, got %d", len(lines))
	}

	// Should have pruning notice
	if !strings.Contains(output, "pruned") {
		t.Errorf("output should contain pruning notice, got:\n%s", output)
	}
}

func TestFormatCodemap_NoLimitUsesDefault(t *testing.T) {
	// With LineLimit = 0, should use DefaultLineLimit (1000)
	files := makeTestFiles(5, 10) // 5 files, 10 symbols each = 60 lines

	output := FormatCodemap(files, FormatOptions{
		LineLimit: 0, // Should use DefaultLineLimit
	})

	// Should NOT have pruning notice since 60 < 1000
	if strings.Contains(output, "pruned") {
		t.Errorf("output should not contain pruning notice for small outputs")
	}
}

func TestFormatCodemap_FilterOverridesSkip(t *testing.T) {
	files := []FileIndex{
		{Path: "vendor/lib.go", Language: "go", Symbols: makeSymbols(5)},
		{Path: "main.go", Language: "go", Symbols: makeSymbols(5)},
	}

	// Without filter, vendor should be skipped
	output := FormatCodemap(files, FormatOptions{
		SkipPatterns: []string{"vendor"},
	})
	if !strings.Contains(output, "skipped by default") {
		t.Errorf("vendor should be skipped")
	}

	// With filter on vendor, it should be included
	output = FormatCodemap(files, FormatOptions{
		SkipPatterns: []string{"vendor"},
		Filter:       "vendor",
	})
	if strings.Contains(output, "skipped") {
		t.Errorf("vendor should NOT be skipped when filtered")
	}
	if !strings.Contains(output, "vendor/lib.go") {
		t.Errorf("vendor/lib.go should be in output")
	}
}

func TestMatchesFilter(t *testing.T) {
	tests := []struct {
		filePath string
		filter   string
		expected bool
	}{
		{"cmd/main.go", "cmd", true},
		{"cmd/main.go", "cmd/", true},
		{"cmd/main.go", "cmd/main.go", true},
		{"cmd/sub/main.go", "cmd", true},
		{"pkg/main.go", "cmd", false},
		{"cmdx/main.go", "cmd", false},
		{"./cmd/main.go", "cmd", true},
		{"cmd/main.go", "./cmd", true},
	}

	for _, tt := range tests {
		t.Run(tt.filePath+"_"+tt.filter, func(t *testing.T) {
			got := matchesFilter(tt.filePath, tt.filter)
			if got != tt.expected {
				t.Errorf("matchesFilter(%q, %q) = %v, want %v",
					tt.filePath, tt.filter, got, tt.expected)
			}
		})
	}
}

func TestIsSkipped(t *testing.T) {
	patterns := []string{"vendor", "internal/gen"}

	tests := []struct {
		filePath string
		expected bool
	}{
		{"vendor/lib.go", true},
		{"vendor/sub/lib.go", true},
		{"internal/gen/types.go", true},
		{"internal/gen/sub/types.go", true},
		{"internal/other/types.go", false},
		{"vendorx/lib.go", false},
		{"main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			got := isSkipped(tt.filePath, patterns)
			if got != tt.expected {
				t.Errorf("isSkipped(%q, patterns) = %v, want %v",
					tt.filePath, got, tt.expected)
			}
		})
	}
}
