# AGENTS.md

This document provides guidance for AI coding agents working on this codebase.

## Project Overview

**go-indexer-mcp** is an MCP (Model Context Protocol) server that indexes Go codebases.
It parses Go source files using the `go/ast` package and extracts symbols (functions,
types, constants, variables) with their line ranges to provide codebase maps for AI assistants.

### Project Structure

```
go-indexer-mcp/
├── main.go       # CLI entry point, command routing (mcp|map subcommands)
├── indexer.go    # Core Go AST parsing and symbol extraction logic
├── mcp.go        # MCP server implementation and tool handlers
├── go.mod        # Module definition (github.com/roveo/go-indexer-mcp)
└── go.sum        # Dependency checksums
```

## Build Commands

```bash
# Build the binary
go build

# Run directly without building
go run .

# Install to $GOPATH/bin
go install

# Format code
make format

# Lint code (with auto-fix)
make lint

# Run all tests
make test
```

## Running the Application

```bash
# Run as MCP server (default mode, communicates via stdio)
./go-indexer-mcp mcp
# or simply
./go-indexer-mcp

# Index a directory and print to stdout
./go-indexer-mcp map [path]
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

- `github.com/modelcontextprotocol/go-sdk` - MCP protocol SDK for Go
- Standard library: `go/ast`, `go/parser`, `go/token` for AST parsing

## Code Style Guidelines

### Import Organization

Group imports in this order, separated by blank lines:
1. Standard library packages
2. External packages

```go
import (
    "context"
    "fmt"
    "os"
    "path/filepath"

    "github.com/modelcontextprotocol/go-sdk/mcp"
)
```

### Naming Conventions

- **Exported types**: PascalCase (e.g., `Symbol`, `FileIndex`, `PackageIndex`)
- **Unexported functions**: camelCase (e.g., `firstLineOfComment`, `symbolKind`)
- **Exported functions**: PascalCase (e.g., `IndexDirectory` if it were exported)
- **Variables**: camelCase for local variables (e.g., `relPath`, `fset`)
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
   f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
   if err != nil {
       return nil  // Skip files that can't be parsed
   }
   ```

### Struct Tags

Use JSON tags with `omitempty` for optional fields. Add `jsonschema` tags for MCP tool inputs:

```go
type Symbol struct {
    Name      string `json:"name"`
    Kind      string `json:"kind"`
    Receiver  string `json:"receiver,omitempty"`
    StartLine int    `json:"start_line"`
    EndLine   int    `json:"end_line"`
    DocHead   string `json:"doc_head,omitempty"`
}

type IndexToolInput struct {
    Path string `json:"path,omitempty" jsonschema:"description=Directory path to index."`
}
```

### Function Documentation

Add doc comments before exported functions:

```go
// indexDirectory walks the directory and indexes all Go files
func indexDirectory(dir string) ([]FileIndex, error) {
```

### Code Patterns

**Type switches** for AST node handling:
```go
switch decl := d.(type) {
case *ast.FuncDecl:
    // handle function
case *ast.GenDecl:
    // handle general declaration
}
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
    if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
        return filepath.SkipDir
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
- Non-Go files

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

## Key Types

| Type | File | Purpose |
|------|------|---------|
| `Symbol` | indexer.go | Represents a code symbol with line range info |
| `FileIndex` | indexer.go | Index of a single Go file (imports + symbols) |
| `PackageIndex` | indexer.go | Index of a package (reserved for future use) |
| `IndexToolInput` | mcp.go | MCP tool input schema |

## Key Functions

| Function | File | Purpose |
|----------|------|---------|
| `indexDirectory` | indexer.go | Walk directory and parse all Go files |
| `collectSymbols` | indexer.go | Extract symbols from AST |
| `formatCompact` | indexer.go | Format index as human-readable text |
| `runMCPServer` | mcp.go | Start MCP server on stdio |
| `indexHandler` | mcp.go | MCP tool handler for `index` command |
