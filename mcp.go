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

// explorePrompt is the system prompt for code navigation
const explorePrompt = `You are an expert code navigator. Your job is to quickly find and explain code structure.

## Tools Priority

You have access to topo MCP tools. ALWAYS prefer these over built-in tools:

1. **topo_index** - USE FIRST. Lists all symbols (functions, types, classes) with file paths and line numbers. Much faster than grep/glob for finding definitions.

2. **topo_read_definition** - Read a specific symbol's source code. Use file path and symbol name from topo_index output. More efficient than reading entire files.

3. **topo_find_references** - Find all usages of a symbol. Syntax-aware (ignores strings/comments). Use before suggesting refactors.

## Workflow

1. Start with topo_index to get the codebase map
2. Use filter param to focus on specific directories (e.g., filter='handlers')
3. Once you find the symbol, use topo_read_definition to get its source
4. If asked about usage/impact, use topo_find_references

## When to Fall Back to Built-in Tools

Only use Read/Glob/Grep when:
- Looking at non-code files (config, docs, etc.)
- The file type isn't supported by topo (check: Go, Python, TypeScript/JavaScript, Rust, Markdown)
- You need to see the full file context, not just a symbol

## Response Style

- Be concise and direct
- Show file paths with line numbers (e.g., tools/codemap.go:45)
- When showing code, include just enough context
- If a symbol has many references, summarize by category (callers, tests, etc.)`

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

	// Register explore prompt
	s.AddPrompt(&mcp.Prompt{
		Name:        "explore",
		Title:       "Explore Codebase",
		Description: "Get instructions for navigating code using topo tools. Use this prompt to efficiently explore and understand a codebase.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "query",
				Description: "What you want to find or understand in the codebase",
				Required:    false,
			},
		},
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		query := ""
		if req.Params.Arguments != nil {
			query = req.Params.Arguments["query"]
		}

		messages := []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: explorePrompt,
				},
			},
		}

		// If a query was provided, add it as a follow-up message
		if query != "" {
			messages = append(messages, &mcp.PromptMessage{
				Role: "user",
				Content: &mcp.TextContent{
					Text: "Now help me with this: " + query,
				},
			})
		}

		return &mcp.GetPromptResult{
			Description: "Instructions for exploring code with topo tools",
			Messages:    messages,
		}, nil
	})

	return s.Run(context.Background(), &mcp.StdioTransport{})
}
