package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/roveo/topo-mcp/gitignore"
	"github.com/roveo/topo-mcp/languages"
	sitter "github.com/smacker/go-tree-sitter"
)

// FindReferencesInput is the input schema for the find_references tool
type FindReferencesInput struct {
	Path   string `json:"path,omitempty" jsonschema_description:"Directory to search in. Defaults to current working directory."`
	Symbol string `json:"symbol" jsonschema_description:"Name of the symbol to find references for."`
}

// FindReferencesTool creates the find_references MCP tool
func FindReferencesTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "find_references",
		Description: `Find all usages of a symbol across the codebase.

Syntax-aware: only finds actual code references, not strings or comments. Better than grep for code navigation.

Use before refactoring to see what would be affected, or to understand how a function/type is used.`,
	}
}

// FindReferencesHandler handles the find_references tool invocation
func FindReferencesHandler(cfg *Config) func(context.Context, *mcp.CallToolRequest, FindReferencesInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input FindReferencesInput) (*mcp.CallToolResult, any, error) {
		if input.Symbol == "" {
			return nil, nil, fmt.Errorf("symbol name is required")
		}

		dir := input.Path
		if dir == "" {
			var err error
			dir, err = os.Getwd()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get working directory: %w", err)
			}
		}

		// Make path absolute if relative
		if !filepath.IsAbs(dir) {
			cwd, err := os.Getwd()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get working directory: %w", err)
			}
			dir = filepath.Join(cwd, dir)
		}

		refs, err := FindReferences(dir, input.Symbol)
		if err != nil {
			return nil, nil, err
		}

		if len(refs) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("No references found for %q", input.Symbol)},
				},
			}, nil, nil
		}

		// Format output
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# References to %q (%d found)\n\n", input.Symbol, len(refs)))

		currentFile := ""
		for _, ref := range refs {
			if ref.File != currentFile {
				if currentFile != "" {
					sb.WriteString("\n")
				}
				sb.WriteString(fmt.Sprintf("## %s\n", ref.File))
				currentFile = ref.File
			}
			sb.WriteString(fmt.Sprintf("  [%d:%d] %s\n", ref.Line, ref.Column, ref.Context))
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: sb.String()},
			},
		}, nil, nil
	}
}

// Reference represents a single reference to a symbol
type Reference struct {
	File    string // Relative file path
	Line    int    // 1-based line number
	Column  int    // 1-based column number
	Context string // The line of code containing the reference
}

// FindReferences finds all references to a symbol in a directory
func FindReferences(dir string, symbolName string) ([]Reference, error) {
	var refs []Reference

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
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Find references in this file
		fileRefs, err := findReferencesInFile(content, symbolName, lang)
		if err != nil {
			return nil // Skip files that can't be parsed
		}

		// Add file path to references
		for i := range fileRefs {
			fileRefs[i].File = relPath
		}

		refs = append(refs, fileRefs...)
		return nil
	})

	return refs, err
}

// findReferencesInFile finds all references to a symbol in a single file
func findReferencesInFile(content []byte, symbolName string, lang languages.Language) ([]Reference, error) {
	// Check if language supports tree-sitter
	tsLang, ok := lang.(languages.TreeSitterLanguage)
	if !ok {
		return nil, fmt.Errorf("language %s doesn't support tree-sitter", lang.Name())
	}

	// Parse the file
	parser := sitter.NewParser()
	parser.SetLanguage(tsLang.TreeSitterLang())
	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	lines := strings.Split(string(content), "\n")
	var refs []Reference

	// Walk the tree looking for identifiers matching the symbol name
	var walk func(node *sitter.Node)
	walk = func(node *sitter.Node) {
		if node == nil {
			return
		}

		// Check if this is an identifier-like node
		if isIdentifierNode(node, lang.Name()) {
			name := node.Content(content)
			if name == symbolName {
				line := int(node.StartPoint().Row)
				col := int(node.StartPoint().Column)

				// Get context (the line of code)
				context := ""
				if line < len(lines) {
					context = strings.TrimSpace(lines[line])
					// Truncate long lines
					if len(context) > 100 {
						context = context[:97] + "..."
					}
				}

				refs = append(refs, Reference{
					Line:    line + 1, // Convert to 1-based
					Column:  col + 1,
					Context: context,
				})
			}
		}

		// Recurse into children
		for i := 0; i < int(node.ChildCount()); i++ {
			walk(node.Child(i))
		}
	}

	walk(tree.RootNode())
	return refs, nil
}

// isIdentifierNode checks if a node is an identifier in the given language
func isIdentifierNode(node *sitter.Node, langName string) bool {
	nodeType := node.Type()

	switch langName {
	case "go":
		return nodeType == "identifier" ||
			nodeType == "type_identifier" ||
			nodeType == "field_identifier" ||
			nodeType == "package_identifier"
	case "python":
		return nodeType == "identifier"
	case "typescript", "javascript":
		return nodeType == "identifier" ||
			nodeType == "property_identifier" ||
			nodeType == "type_identifier"
	case "rust":
		return nodeType == "identifier" ||
			nodeType == "type_identifier" ||
			nodeType == "field_identifier"
	default:
		return nodeType == "identifier"
	}
}
