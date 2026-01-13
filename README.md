# Topo MCP

An MCP (Model Context Protocol) server providing **code topology tools** for LLMs - helping them understand and navigate codebases.

## What is Topo?

LLMs need to know *where things are* in a codebase before they can jump to them. Topo provides the map.

Traditional developer tools like LSP were designed for IDEs with cursors and real-time feedback. LLMs operate differently - they work with context windows, batch processing, and need high-level structural understanding.

Topo provides tools that give LLMs what they need:
- **`index`** - Map the terrain (list all symbols with locations)
- **`read_definition`** - Jump to a symbol and read its code
- **`write_definition`** - Replace a symbol's code
- **`find_references`** - Find everywhere a symbol is used

## Example Output

```
## src/server.go
  type Server struct [15-42] // Server handles HTTP requests
  NewServer(*Config) *Server [44-60]
  (*Server) Start(context.Context) error [62-85]
  (*Server) handleRequest(http.ResponseWriter, *http.Request) [87-120]

## src/config.go
  type Config struct [8-15] // Config holds server settings
  LoadConfig(string) (*Config, error) [17-35]
```

## Supported Languages

| Language | Extensions | Build Tag |
|----------|------------|-----------|
| Go | `.go` | `lang_go` |
| Python | `.py` | `lang_python` |
| TypeScript | `.ts`, `.tsx` | `lang_typescript` |
| JavaScript | `.js`, `.jsx`, `.mjs`, `.cjs` | `lang_typescript` |
| Rust | `.rs` | `lang_rust` |

## Installation

### Pre-built Profiles

Choose a build profile that matches your stack:

| Profile | Languages | Binary |
|---------|-----------|--------|
| Go only | Go | `topo-go` |
| Python only | Python | `topo-python` |
| TypeScript/JS | TS/JS | `topo-typescript` |
| Rust only | Rust | `topo-rust` |
| Backend | Go, Python, Rust | `topo-backend` |
| Frontend | TypeScript/JavaScript | `topo-frontend` |
| Fullstack | Go, TypeScript/JS | `topo-fullstack` |
| Web | Python, TypeScript/JS | `topo-web` |
| ML | Python, Rust | `topo-ml` |
| All | All languages | `topo-all` |

### Build from Source

```bash
git clone https://github.com/roveo/topo-mcp.git
cd topo-mcp

# Build default (all languages)
make build

# Build with specific language(s) only
make build-go
go build -tags lang_go .
go build -tags "lang_go,lang_python" .

# Build all profiles
make build-profiles
```

### Install with Go

```bash
# All languages (default)
go install github.com/roveo/topo-mcp@latest

# Specific language(s) only
go install -tags lang_go github.com/roveo/topo-mcp@latest
go install -tags "lang_go,lang_python" github.com/roveo/topo-mcp@latest
```

## Usage

### MCP Server Mode

Run as an MCP server (communicates via stdio):

```bash
topo mcp
```

### Available Tools

#### `index`
List all symbols in a codebase with their line ranges. Output is limited to 1000 lines by default to keep responses manageable; large directories are automatically pruned with a notice showing which directories were omitted.

| Parameter | Description |
|-----------|-------------|
| `path` | Directory to index (default: cwd) |
| `filter` | Path filter to show only matching files/directories |

#### `read_definition`
Get the source code of a symbol by name and file path.

| Parameter | Description |
|-----------|-------------|
| `file` | Relative file path |
| `symbol` | Name of the symbol to read |

#### `write_definition`
Replace a symbol's source code.

| Parameter | Description |
|-----------|-------------|
| `file` | Relative file path |
| `symbol` | Name of the symbol to replace |
| `code` | New source code for the symbol |

#### `find_references`
Find all references to a symbol across the codebase.

| Parameter | Description |
|-----------|-------------|
| `path` | Directory to search (default: cwd) |
| `symbol` | Name of the symbol to find |

### CLI Mode

```bash
# Index current directory
topo map

# Index a specific directory
topo map /path/to/project

# Filter to specific package/directory
topo map --filter src/handlers

# Skip certain paths by default
topo map --skip generated --skip vendor

# Limit output lines (default: 1000, 0 = no limit)
topo map --limit 500
```

### MCP Client Configuration

#### OpenCode

Add to your `opencode.json`:

```json
{
  "mcp": {
    "topo": {
      "type": "local",
      "command": ["topo", "mcp"]
    }
  }
}
```

With skip patterns:

```json
{
  "mcp": {
    "topo": {
      "type": "local",
      "command": ["topo", "mcp", "--skip", "generated", "--skip", "testdata"]
    }
  }
}
```

#### Claude Code

Add to your Claude Code MCP settings (`~/.claude/claude_desktop_config.json` or via `claude mcp add`):

```json
{
  "mcpServers": {
    "topo": {
      "command": "topo",
      "args": ["mcp"]
    }
  }
}
```

Or use the CLI:

```bash
claude mcp add topo -- topo mcp
```

With skip patterns:

```bash
claude mcp add topo -- topo mcp --skip generated --skip testdata
```

#### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "topo": {
      "command": "topo",
      "args": ["mcp"]
    }
  }
}
```

## Output Format Examples

### Go
```
## main.go
  main() [10-15]
  type Config struct [17-22] // Config holds settings
  (*Config) Validate() error [24-30]
  const DefaultTimeout [32]
  var ErrNotFound [34]
```

### Python
```
## app.py
  VERSION [5]
  @dataclass class Config [8-15] // Application configuration
  class Server(BaseServer) [17-45] // HTTP server implementation
  def main(args: List[str]) -> int [47-60] // Entry point
```

### TypeScript
```
## server.ts
  interface Config [5-12] // Server configuration
  type Handler [14-16]
  class Server extends EventEmitter [18-50] // Main server class
  async function startServer(config: Config): Promise<void> [52-70]
```

### Rust
```
## lib.rs
  pub mod server [3]
  pub struct Config [5-12] // Server configuration
  pub enum Error [14-20]
  pub trait Handler [22-28]
  impl Config: pub fn new() -> Self [30-35]
```

## Automatic Exclusions

The indexer automatically skips:
- Hidden directories (`.git`, `.vscode`, etc.)
- `vendor/` directory
- `node_modules/` directory

## Architecture

```
topo-mcp/
├── languages/
│   ├── language.go      # Symbol interface, Range, Position
│   ├── registry.go      # Language registry
│   ├── golang/          # Go parser (tree-sitter)
│   ├── python/          # Python parser (tree-sitter)
│   ├── typescript/      # TS/JS parser (tree-sitter)
│   └── rust/            # Rust parser (tree-sitter)
├── tools/
│   ├── codemap.go       # index tool
│   ├── read_definition.go
│   ├── write_definition.go
│   └── find_references.go
├── mcp.go               # MCP server implementation
└── main.go              # CLI entry point
```

Each language is in its own package with build tags, allowing compile-time selection of which languages to include.

## Dependencies

- [tree-sitter](https://tree-sitter.github.io/) via [go-tree-sitter](https://github.com/smacker/go-tree-sitter) - Fast, incremental parsing
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) - Model Context Protocol
- [Cobra](https://github.com/spf13/cobra) - CLI framework

## License

MIT
