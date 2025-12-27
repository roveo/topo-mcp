package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// serverSkipPatterns holds skip patterns configured at server startup
var serverSkipPatterns []string

// IndexToolInput is the input schema for the index tool
type IndexToolInput struct {
	Path   string `json:"path,omitempty" jsonschema_description:"Directory path to index. Defaults to current working directory if not specified."`
	Filter string `json:"filter,omitempty" jsonschema_description:"Optional path filter to show only a specific package (directory) or file. When specified, only files matching this prefix will have their symbols shown. Use this to get a compact map of just the relevant part of the codebase. Overrides any default skip patterns for matching files."`
}

func indexHandler(ctx context.Context, req *mcp.CallToolRequest, input IndexToolInput) (*mcp.CallToolResult, any, error) {
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

	files, err := indexDirectory(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to index directory: %w", err)
	}

	output := formatCompact(files, FormatOptions{
		SkipPatterns: serverSkipPatterns,
		Filter:       input.Filter,
	})
	if output == "" {
		output = "No Go symbols found in the specified directory."
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: output},
		},
	}, nil, nil
}

func runMap(path string, skipPatterns []string, filter string) error {
	// Make path absolute if relative
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		path = filepath.Join(cwd, path)
	}

	files, err := indexDirectory(path)
	if err != nil {
		return fmt.Errorf("failed to index directory: %w", err)
	}

	output := formatCompact(files, FormatOptions{
		SkipPatterns: skipPatterns,
		Filter:       filter,
	})
	if output == "" {
		output = "No Go symbols found in the specified directory."
	}

	fmt.Print(output)
	return nil
}

func runMCPServer(skipPatterns []string) error {
	serverSkipPatterns = skipPatterns

	s := mcp.NewServer(&mcp.Implementation{
		Name:    "go-indexer",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "index",
		Description: "Index a Go codebase and return a compact listing of all symbols (functions, types, consts, vars) with their line ranges.",
	}, indexHandler)

	return s.Run(context.Background(), &mcp.StdioTransport{})
}
