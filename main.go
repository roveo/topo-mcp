package main

import (
	"fmt"
	"os"

	"github.com/roveo/topo-mcp/tools"
	"github.com/spf13/cobra"
)

var skipPatterns []string
var lineLimit int

var rootCmd = &cobra.Command{
	Use:   "topo",
	Short: "Code topology tools for LLMs",
	Long: `topo is an MCP (Model Context Protocol) server providing code navigation tools for LLMs.
It parses source files and provides tools to index symbols, read/write definitions,
and find references across codebases. Supports Go, Python, TypeScript/JavaScript, and Rust.`,
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run as MCP server (communicates via stdio)",
	Long: `Run as an MCP server that communicates via stdio.
Exposes tools: index, read_definition, write_definition, find_references.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMCPServer(skipPatterns, lineLimit)
	},
}

var mapCmd = &cobra.Command{
	Use:   "map [path]",
	Short: "Index a directory and print the map to stdout",
	Long: `Index a codebase directory and print a compact listing of all symbols
(functions, types, classes, etc.) with their line ranges to stdout.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		filter, _ := cmd.Flags().GetString("filter")
		return runMap(path, skipPatterns, filter, lineLimit)
	},
}

func init() {
	// Add --skip flag to root (inherited by all subcommands)
	rootCmd.PersistentFlags().StringArrayVar(&skipPatterns, "skip", nil,
		"Path prefixes to skip by default (can be specified multiple times)")

	// Add --limit flag to root (inherited by all subcommands)
	rootCmd.PersistentFlags().IntVar(&lineLimit, "limit", tools.DefaultLineLimit,
		"Maximum lines in output (0 = no limit)")

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
