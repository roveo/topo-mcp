//go:build !lang_go && !lang_python && !lang_typescript && !lang_rust

package main

// Import all language packages by default (when no lang_* tags specified)
import (
	_ "github.com/roveo/topo-mcp/languages/golang"
	_ "github.com/roveo/topo-mcp/languages/python"
	_ "github.com/roveo/topo-mcp/languages/rust"
	_ "github.com/roveo/topo-mcp/languages/typescript"
)
