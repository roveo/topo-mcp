package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ReadDefinitionInput is the input schema for the read_definition tool
type ReadDefinitionInput struct {
	File   string `json:"file" jsonschema_description:"Relative file path from the project root (e.g., 'cmd/main.go', 'src/utils.py')."`
	Symbol string `json:"symbol" jsonschema_description:"Name of the symbol to retrieve (function, type, class, method, etc.). For methods, use just the method name without the receiver."`
}

// ReadDefinitionTool creates the read_definition MCP tool
func ReadDefinitionTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "read_definition",
		Description: "Get the source code of a symbol (function, type, class, etc.) by name and file path. Similar to LSP's 'Go to Definition'. Returns the complete source code of the symbol including its signature and body.",
	}
}

// ReadDefinitionHandler handles the read_definition tool invocation
func ReadDefinitionHandler(cfg *Config) func(context.Context, *mcp.CallToolRequest, ReadDefinitionInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ReadDefinitionInput) (*mcp.CallToolResult, any, error) {
		if input.File == "" {
			return nil, nil, fmt.Errorf("file path is required")
		}
		if input.Symbol == "" {
			return nil, nil, fmt.Errorf("symbol name is required")
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

		// Find the symbol and get its source code
		symbol, lines, err := FindSymbol(filePath, input.Symbol)
		if err != nil {
			return nil, nil, err
		}

		// Format the output
		loc := symbol.Location()
		startLine := loc.Start.Line + 1 // Convert to 1-based
		endLine := loc.End.Line + 1

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# %s in %s [%d-%d]\n\n", symbol.String(), input.File, startLine, endLine))

		// Add doc comment if available
		if doc, ok := symbol.(interface{ DocComment() string }); ok {
			if docStr := doc.DocComment(); docStr != "" {
				sb.WriteString(fmt.Sprintf("// %s\n\n", docStr))
			}
		}

		sb.WriteString("```\n")
		for i, line := range lines {
			// Show line numbers
			lineNum := startLine + i
			sb.WriteString(fmt.Sprintf("%4d | %s\n", lineNum, line))
		}
		sb.WriteString("```\n")

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: sb.String()},
			},
		}, nil, nil
	}
}
