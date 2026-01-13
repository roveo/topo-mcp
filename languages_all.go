//go:build lang_all

package main

// Import all language packages to register them
import (
	_ "github.com/roveo/topo-mcp/languages/golang"
	_ "github.com/roveo/topo-mcp/languages/python"
	_ "github.com/roveo/topo-mcp/languages/rust"
	_ "github.com/roveo/topo-mcp/languages/typescript"
)
