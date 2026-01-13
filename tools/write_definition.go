package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// WriteDefinitionInput is the input schema for the write_definition tool
type WriteDefinitionInput struct {
	File   string `json:"file" jsonschema_description:"Relative file path from the project root (e.g., 'cmd/main.go', 'src/utils.py')."`
	Symbol string `json:"symbol" jsonschema_description:"Name of the symbol to replace (function, type, class, method, etc.). For methods, use just the method name without the receiver."`
	Code   string `json:"code" jsonschema_description:"The new source code for the symbol. Should be complete and valid code that replaces the entire symbol definition."`
}

// WriteDefinitionTool creates the write_definition MCP tool
func WriteDefinitionTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "write_definition",
		Description: "Replace the source code of a symbol (function, type, class, etc.) by name and file path. The inverse of read_definition. Replaces the entire symbol definition with the provided code.",
	}
}

// WriteDefinitionHandler handles the write_definition tool invocation
func WriteDefinitionHandler(cfg *Config) func(context.Context, *mcp.CallToolRequest, WriteDefinitionInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input WriteDefinitionInput) (*mcp.CallToolResult, any, error) {
		if input.File == "" {
			return nil, nil, fmt.Errorf("file path is required")
		}
		if input.Symbol == "" {
			return nil, nil, fmt.Errorf("symbol name is required")
		}
		if input.Code == "" {
			return nil, nil, fmt.Errorf("code is required")
		}

		// Make path absolute if relative
		filePath := input.File
		if !filepath.IsAbs(filePath) {
			cwd, err := os.Getwd()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get working directory: %w", err)
			}
			filePath = filepath.Join(cwd, filePath)
		}

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("file not found: %s", input.File)
		}

		// Replace the symbol
		err := ReplaceSymbol(filePath, input.Symbol, input.Code)
		if err != nil {
			return nil, nil, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Successfully replaced %s in %s", input.Symbol, input.File)},
			},
		}, nil, nil
	}
}

// ReplaceSymbol replaces a symbol's source code in a file
func ReplaceSymbol(filePath string, symbolName string, newCode string) error {
	symbols, err := ParseFile(filePath)
	if err != nil {
		return err
	}

	// Find the symbol
	var found = -1
	for i, sym := range symbols {
		if sym.Name() == symbolName {
			found = i
			break
		}
	}

	if found == -1 {
		return fmt.Errorf("symbol %q not found in %s", symbolName, filePath)
	}

	symbol := symbols[found]
	loc := symbol.Location()

	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	startLine := loc.Start.Line
	endLine := loc.End.Line

	// Bounds check
	if startLine < 0 {
		startLine = 0
	}
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	// Build new content: lines before + new code + lines after
	var newLines []string
	newLines = append(newLines, lines[:startLine]...)

	// Add new code (split into lines, trim trailing newline to avoid double)
	newCode = strings.TrimSuffix(newCode, "\n")
	newLines = append(newLines, strings.Split(newCode, "\n")...)

	// Add lines after the symbol
	if endLine+1 < len(lines) {
		newLines = append(newLines, lines[endLine+1:]...)
	}

	// Write back
	newContent := strings.Join(newLines, "\n")
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
