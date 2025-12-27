package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var skipPatterns []string

var rootCmd = &cobra.Command{
	Use:   "go-indexer-mcp",
	Short: "MCP server that indexes Go codebases",
	Long: `go-indexer-mcp is an MCP (Model Context Protocol) server that indexes Go codebases.
It parses Go source files and extracts symbols (functions, types, constants, variables)
with their line ranges to provide codebase maps for AI assistants.`,
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run as MCP server (communicates via stdio)",
	Long: `Run as an MCP server that communicates via stdio.
The server exposes an 'index' tool that can be called to index Go codebases.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMCPServer(skipPatterns)
	},
}

var mapCmd = &cobra.Command{
	Use:   "map [path]",
	Short: "Index a directory and print the map to stdout",
	Long: `Index a Go codebase directory and print a compact listing of all symbols
(functions, types, consts, vars) with their line ranges to stdout.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		filter, _ := cmd.Flags().GetString("filter")
		return runMap(path, skipPatterns, filter)
	},
}

func init() {
	// Add --skip flag to root (inherited by all subcommands)
	rootCmd.PersistentFlags().StringArrayVar(&skipPatterns, "skip", nil,
		"Path prefixes to skip by default (can be specified multiple times)")

	// Add --filter flag to map command
	mapCmd.Flags().StringP("filter", "f", "",
		"Only show symbols for files matching this path prefix (file or directory)")

	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(mapCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
