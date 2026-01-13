package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/roveo/topo-mcp/tools"
)

// serverConfig holds the server configuration
var serverConfig *tools.Config

func runMap(path string, skipPatterns []string, filter string, lineLimit int) error {
	// Make path absolute if relative
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		path = filepath.Join(cwd, path)
	}

	files, err := tools.IndexDirectory(path)
	if err != nil {
		return fmt.Errorf("failed to index directory: %w", err)
	}

	output := tools.FormatCodemap(files, tools.FormatOptions{
		SkipPatterns: skipPatterns,
		Filter:       filter,
		LineLimit:    lineLimit,
	})
	if output == "" {
		output = "No symbols found in the specified directory."
	}

	fmt.Print(output)
	return nil
}

func runMCPServer(skipPatterns []string, lineLimit int) error {
	serverConfig = &tools.Config{
		SkipPatterns: skipPatterns,
		LineLimit:    lineLimit,
	}

	s := mcp.NewServer(&mcp.Implementation{
		Name:    "topo",
		Version: "1.0.0",
	}, nil)

	// Register codemap tool
	mcp.AddTool(s, tools.CodemapTool(), tools.CodemapHandler(serverConfig))

	// Register read_definition tool
	mcp.AddTool(s, tools.ReadDefinitionTool(), tools.ReadDefinitionHandler(serverConfig))

	// Register write_definition tool
	mcp.AddTool(s, tools.WriteDefinitionTool(), tools.WriteDefinitionHandler(serverConfig))

	// Register find_references tool
	mcp.AddTool(s, tools.FindReferencesTool(), tools.FindReferencesHandler(serverConfig))

	return s.Run(context.Background(), &mcp.StdioTransport{})
}
