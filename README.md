# go-indexer-mcp

An MCP (Model Context Protocol) server that indexes Go codebases. It parses Go source files using the `go/ast` package and extracts symbols (functions, types, constants, variables) with their line ranges to provide codebase maps for AI assistants.

## Features

- Extracts all Go symbols: functions, methods, types, constants, and variables
- Provides line ranges for each symbol for easy navigation
- Includes function signatures and type kinds (struct, interface, etc.)
- Extracts doc comments (first line) for types and functions
- Supports filtering by file or directory path
- Configurable skip patterns to exclude directories by default
- Works as an MCP server or standalone CLI tool

## Installation

```bash
go install github.com/roveo/go-indexer-mcp@latest
```

Or build from source:

```bash
git clone https://github.com/roveo/go-indexer-mcp.git
cd go-indexer-mcp
go build
```

## Usage

### MCP Server Mode

Run as an MCP server (default mode, communicates via stdio):

```bash
go-indexer-mcp mcp
```

The server exposes an `index` tool with the following parameters:

| Parameter | Description |
|-----------|-------------|
| `path` | Directory path to index. Defaults to current working directory. |
| `filter` | Optional path filter to show only a specific package or file. Overrides skip patterns. |

### CLI Mode

Index a directory and print the map to stdout:

```bash
go-indexer-mcp map [path]
```

Options:
- `-f, --filter`: Only show symbols for files matching this path prefix
- `--skip`: Path prefixes to skip by default (can be specified multiple times)

### Examples

Index the current directory:
```bash
go-indexer-mcp map
```

Index a specific directory:
```bash
go-indexer-mcp map /path/to/project
```

Index only a specific package:
```bash
go-indexer-mcp map --filter cmd/server
```

Skip parts of the code by default:
```bash
go-indexer-mcp map --skip generated --skip internal/mocks
```

## Output Format

The output is a compact, human-readable format:

```
## main.go
  main() [10-15]
  type Config struct [17-22] // Config holds application settings
  (*Config) Validate() error [24-30]
  const DefaultTimeout [32]
  var ErrNotFound [34]

## pkg/handler/handler.go
  type Handler interface [5-12] // Handler processes requests
  NewHandler(*Config) Handler [14-20]
```

Each symbol shows:
- **Functions**: `name(params) returns [start-end]`
- **Methods**: `(receiver) name(params) returns [start-end]`
- **Types**: `type Name kind [start-end]`
- **Consts/Vars**: `const/var Name [line]`
- Doc comments are appended with `// ...` when present

## MCP Configuration

Add to your MCP client configuration:

```json
{
  "mcpServers": {
    "go-indexer": {
      "command": "go-indexer-mcp",
      "args": ["mcp"]
    }
  }
}
```

With skip patterns:

```json
{
  "mcpServers": {
    "go-indexer": {
      "command": "go-indexer-mcp",
      "args": ["mcp", "--skip", "generated", "--skip", "testdata"]
    }
  }
}
```

## Directories Skipped Automatically

The indexer automatically skips:
- Hidden directories (starting with `.`)
- `vendor/` directory
- `node_modules/` directory

## Dependencies

- [github.com/modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) - MCP protocol SDK for Go
- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- Standard library: `go/ast`, `go/parser`, `go/token` for AST parsing

## License

MIT
