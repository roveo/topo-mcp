//go:build lang_go || (!lang_all && !lang_python && !lang_typescript && !lang_rust)

package main

// Import Go language package to register it (default when no tags specified)
import (
	_ "github.com/roveo/topo-mcp/languages/golang"
)
