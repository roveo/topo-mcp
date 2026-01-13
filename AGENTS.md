# AGENTS.md

This document provides guidance for AI coding agents working on this codebase.

## Project Overview

**topo-mcp** is an MCP (Model Context Protocol) server providing code topology tools for LLMs.
It parses source files using tree-sitter and provides tools to index symbols, read/write definitions,
and find references across codebases. Supports Go, Python, TypeScript/JavaScript, and Rust.

### Project Structure

```
topo-mcp/
├── main.go              # CLI entry point, command routing (mcp|map subcommands)
├── mcp.go               # MCP server implementation and tool registration
├── languages/
│   ├── language.go      # Symbol interface, Range, Position types
│   ├── registry.go      # Language registry (extension → parser mapping)
│   ├── util.go          # NodeRange helper for tree-sitter
│   ├── golang/          # Go parser (build tag: lang_go)
│   ├── python/          # Python parser (build tag: lang_python)
│   ├── typescript/      # TS/JS parser (build tag: lang_typescript)
│   └── rust/            # Rust parser (build tag: lang_rust)
├── tools/
│   ├── tools.go         # Shared types (Config, FileIndex) and utilities
│   ├── codemap.go       # index tool - list all symbols
│   ├── read_definition.go   # read_definition tool - get symbol source
│   ├── write_definition.go  # write_definition tool - replace symbol source
│   └── find_references.go   # find_references tool - find symbol usages
├── languages_*.go       # Build tag files for language registration
├── go.mod               # Module: github.com/roveo/topo-mcp
└── go.sum
```

## Build Commands

```bash
# Build the binary (all languages, default)
make build

# Build with specific language(s) only
make build-go
go build -tags lang_go .
go build -tags "lang_go,lang_python" .

# Run directly without building
go run .

# Install to $GOPATH/bin
go install .

# Format code
make format

# Lint code (with auto-fix)
make lint

# Run all tests
make test
```

## Running the Application

```bash
# Run as MCP server (communicates via stdio)
topo mcp

# Index a directory and print to stdout
topo map [path]

# With options
topo map --filter src/handlers --skip vendor --limit 500
```

## Test Commands

```bash
# Run all tests
make test

# Run a single test by name
go test -mod=readonly -run TestFunctionName ./... -count=1

# Run tests with verbose output
go test -mod=readonly -v ./... -count=1

# Run tests with coverage
go test -mod=readonly -cover ./... -count=1

# Run tests with race detector
go test -mod=readonly -race ./... -count=1
```

**Note:** Test files should be named `*_test.go` and placed alongside the code they test.

## Dependencies

- `github.com/smacker/go-tree-sitter` - Tree-sitter bindings for Go
- `github.com/modelcontextprotocol/go-sdk` - MCP protocol SDK for Go
- `github.com/spf13/cobra` - CLI framework

## Code Style Guidelines

### Import Organization

Group imports in this order, separated by blank lines:
1. Standard library packages
2. External packages
3. Internal packages

```go
import (
    "context"
    "fmt"
    "os"

    "github.com/modelcontextprotocol/go-sdk/mcp"
    sitter "github.com/smacker/go-tree-sitter"

    "github.com/roveo/topo-mcp/languages"
)
```

### Naming Conventions

- **Exported types**: PascalCase (e.g., `Symbol`, `FileIndex`, `Reference`)
- **Unexported functions**: camelCase (e.g., `findReferencesInFile`, `isIdentifierNode`)
- **Exported functions**: PascalCase (e.g., `IndexDirectory`, `FindSymbol`)
- **Variables**: camelCase for local variables (e.g., `relPath`, `content`)
- **Constants**: camelCase for unexported, PascalCase for exported

### Error Handling

1. **Wrap errors with context** using `fmt.Errorf`:
   ```go
   return fmt.Errorf("failed to index directory: %w", err)
   ```

2. **Return errors up the call stack** - don't swallow errors silently

3. **Print errors to stderr** in CLI entry points:
   ```go
   fmt.Fprintf(os.Stderr, "error: %v\n", err)
   os.Exit(1)
   ```

4. **Skip non-fatal errors** when appropriate (e.g., unparseable files):
   ```go
   _, symbols, err := lang.Parse(content)
   if err != nil {
       return nil  // Skip files that can't be parsed
   }
   ```

### Struct Tags

Use JSON tags with `omitempty` for optional fields. Add `jsonschema_description` tags for MCP tool inputs:

```go
type ReadDefinitionInput struct {
    File   string `json:"file" jsonschema_description:"Relative file path from the project root."`
    Symbol string `json:"symbol" jsonschema_description:"Name of the symbol to retrieve."`
}
```

### Function Documentation

Add doc comments before exported functions:

```go
// FindSymbol finds a symbol by name in a file
// Returns the symbol and the file content lines for that symbol
func FindSymbol(filePath string, symbolName string) (languages.Symbol, []string, error) {
```

### Code Patterns

**Tree-sitter node walking**:
```go
var walk func(node *sitter.Node)
walk = func(node *sitter.Node) {
    if node == nil {
        return
    }
    // process node
    for i := 0; i < int(node.ChildCount()); i++ {
        walk(node.Child(i))
    }
}
walk(tree.RootNode())
```

**strings.Builder** for efficient string concatenation:
```go
var sb strings.Builder
fmt.Fprintf(&sb, "## %s\n", file.Path)
return sb.String()
```

**filepath.Walk** for directory traversal with skip logic:
```go
err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
    if info.IsDir() {
        name := info.Name()
        if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" {
            return filepath.SkipDir
        }
    }
    // process file
    return nil
})
```

### Directories to Skip When Indexing

The indexer automatically skips:
- Hidden directories (starting with `.`)
- `vendor/` directory
- `node_modules/` directory

### MCP Server Patterns

MCP tool handlers follow this signature:
```go
func handler(ctx context.Context, req *mcp.CallToolRequest, input InputType) (*mcp.CallToolResult, any, error)
```

Return results as `mcp.TextContent`:
```go
return &mcp.CallToolResult{
    Content: []mcp.Content{
        &mcp.TextContent{Text: output},
    },
}, nil, nil
```

Tool registration in mcp.go:
```go
mcp.AddTool(s, tools.SomeToolTool(), tools.SomeToolHandler(serverConfig))
```

### Adding a New Language

1. Create `languages/newlang/newlang.go` with build tag `//go:build lang_newlang || lang_all`
2. Implement the `Language` interface (Name, Extensions, Parse)
3. Optionally implement `TreeSitterLanguage` for find_references support
4. Register in `init()`: `languages.Register(&Language{})`
5. Create `languages_newlang.go` at root with the import
6. Add build targets to Makefile

### Adding a New Tool

1. Create `tools/new_tool.go` with:
   - `NewToolInput` struct with jsonschema tags
   - `NewToolTool() *mcp.Tool` returning tool metadata
   - `NewToolHandler(cfg *Config)` returning the handler function
2. Add tests in `tools/new_tool_test.go`
3. Register in `mcp.go`: `mcp.AddTool(s, tools.NewToolTool(), tools.NewToolHandler(serverConfig))`

## Key Types

| Type | File | Purpose |
|------|------|---------|
| `languages.Symbol` | languages/language.go | Interface for code symbols |
| `languages.Language` | languages/language.go | Interface for language parsers |
| `tools.Config` | tools/tools.go | Server-wide configuration |
| `tools.FileIndex` | tools/tools.go | Index of a single source file |
| `tools.Reference` | tools/find_references.go | A reference to a symbol |

## Key Functions

| Function | File | Purpose |
|----------|------|---------|
| `tools.IndexDirectory` | tools/tools.go | Walk directory and parse all files |
| `tools.FindSymbol` | tools/tools.go | Find a symbol by name in a file |
| `tools.ReplaceSymbol` | tools/write_definition.go | Replace a symbol's source code |
| `tools.FindReferences` | tools/find_references.go | Find all references to a symbol |
| `tools.FormatCodemap` | tools/codemap.go | Format index as human-readable text |
| `runMCPServer` | mcp.go | Start MCP server on stdio |

## MCP Tools

| Tool | Purpose |
|------|---------|
| `index` | List all symbols in a codebase with line ranges |
| `read_definition` | Get the source code of a symbol |
| `write_definition` | Replace a symbol's source code |
| `find_references` | Find all references to a symbol |
