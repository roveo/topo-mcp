package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CodemapInput is the input schema for the codemap tool
type CodemapInput struct {
	Path   string `json:"path,omitempty" jsonschema_description:"Directory path to index. Defaults to current working directory if not specified."`
	Filter string `json:"filter,omitempty" jsonschema_description:"Optional path filter to show only a specific package (directory) or file. When specified, only files matching this prefix will have their symbols shown. Use this to get a compact map of just the relevant part of the codebase. Overrides any default skip patterns for matching files."`
}

// CodemapTool creates the codemap MCP tool
func CodemapTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "index",
		Description: "Index a codebase and return a compact listing of all symbols (functions, types, classes, etc.) with their line ranges. Supports Go, Python, TypeScript/JavaScript, and Rust (depending on build configuration).",
	}
}

// CodemapHandler handles the codemap tool invocation
func CodemapHandler(cfg *Config) func(context.Context, *mcp.CallToolRequest, CodemapInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CodemapInput) (*mcp.CallToolResult, any, error) {
		dir := input.Path
		if dir == "" {
			var err error
			dir, err = os.Getwd()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get working directory: %w", err)
			}
		}

		// Make path absolute if relative
		if !filepath.IsAbs(dir) {
			cwd, err := os.Getwd()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get working directory: %w", err)
			}
			dir = filepath.Join(cwd, dir)
		}

		files, err := IndexDirectory(dir)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to index directory: %w", err)
		}

		output := FormatCodemap(files, FormatOptions{
			SkipPatterns: cfg.SkipPatterns,
			Filter:       input.Filter,
			LineLimit:    cfg.LineLimit,
		})
		if output == "" {
			output = "No symbols found in the specified directory."
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: output},
			},
		}, nil, nil
	}
}

// FormatOptions controls how the codemap is formatted
type FormatOptions struct {
	SkipPatterns []string // Path prefixes to skip by default
	Filter       string   // If set, only show files matching this prefix (overrides skip)
	LineLimit    int      // Maximum lines in output (0 = no limit, default = DefaultLineLimit)
}

// FormatCodemap formats the index in a compact human-readable format
func FormatCodemap(files []FileIndex, opts FormatOptions) string {
	// Apply line limit if set
	limit := opts.LineLimit
	if limit == 0 {
		limit = DefaultLineLimit
	}

	// Build tree and prune if necessary
	tree := buildDirTree(files, opts)
	prunedFiles, prunedDirs := pruneToLimit(tree, limit)

	var sb strings.Builder

	// Show pruned directories notice
	if len(prunedDirs) > 0 {
		sb.WriteString("# Note: Output pruned to fit line limit\n")
		sb.WriteString("# Pruned directories: ")
		sb.WriteString(strings.Join(prunedDirs, ", "))
		sb.WriteString("\n\n")
	}

	// Handle skipped files (not pruned, but skipped by skip patterns)
	for _, file := range files {
		if opts.Filter != "" {
			continue // Filter overrides skip
		}
		if isSkipped(file.Path, opts.SkipPatterns) {
			sb.WriteString(fmt.Sprintf("## %s\n", file.Path))
			sb.WriteString("  (skipped by default - use filter parameter to index this path explicitly)\n\n")
		}
	}

	// Format pruned files
	for _, file := range prunedFiles {
		if len(file.Symbols) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("## %s\n", file.Path))

		for _, sym := range file.Symbols {
			loc := sym.Location()
			// Convert 0-based to 1-based for display
			startLine := loc.Start.Line + 1
			endLine := loc.End.Line + 1

			var line string
			if startLine == endLine {
				line = fmt.Sprintf("  %s [%d]", sym.String(), startLine)
			} else {
				line = fmt.Sprintf("  %s [%d-%d]", sym.String(), startLine, endLine)
			}

			// Add docstring for types and functions if available
			if doc, ok := sym.(interface{ DocComment() string }); ok {
				if docStr := doc.DocComment(); docStr != "" {
					line += " // " + docStr
				}
			}

			sb.WriteString(line + "\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// matchesFilter checks if a file path matches the filter.
// Supports both exact file match and directory/package prefix match.
func matchesFilter(filePath, filter string) bool {
	// Normalize filter (remove leading ./)
	filter = strings.TrimPrefix(filter, "./")
	filePath = strings.TrimPrefix(filePath, "./")

	// Exact match
	if filePath == filter {
		return true
	}

	// Directory prefix match (filter="cmd" matches "cmd/main.go")
	filterDir := strings.TrimSuffix(filter, "/")
	if strings.HasPrefix(filePath, filterDir+"/") {
		return true
	}

	return false
}

// isSkipped checks if a file path matches any skip pattern (prefix match)
func isSkipped(filePath string, patterns []string) bool {
	filePath = strings.TrimPrefix(filePath, "./")
	for _, pattern := range patterns {
		pattern = strings.TrimPrefix(pattern, "./")
		pattern = strings.TrimSuffix(pattern, "/")
		if filePath == pattern || strings.HasPrefix(filePath, pattern+"/") {
			return true
		}
	}
	return false
}

// fileLineCount returns the number of output lines a file would produce
// Each file contributes: 1 (header) + len(symbols) + 1 (blank line)
func fileLineCount(file FileIndex) int {
	if len(file.Symbols) == 0 {
		return 0
	}
	return 1 + len(file.Symbols) + 1 // header + symbols + blank line
}

// dirNode represents a directory in the tree structure for pruning
type dirNode struct {
	name     string
	path     string              // Full relative path
	files    []FileIndex         // Files directly in this directory
	children map[string]*dirNode // Subdirectories
	lines    int                 // Total lines in this subtree
}

// buildDirTree builds a tree structure from flat file list
func buildDirTree(files []FileIndex, opts FormatOptions) *dirNode {
	root := &dirNode{
		name:     "",
		path:     "",
		children: make(map[string]*dirNode),
	}

	for _, file := range files {
		// Apply filter/skip logic
		if opts.Filter != "" {
			if !matchesFilter(file.Path, opts.Filter) {
				continue
			}
		} else if isSkipped(file.Path, opts.SkipPatterns) {
			// Skipped files still count as 3 lines (header + skip message + blank)
			root.lines += 3
			continue
		}

		if len(file.Symbols) == 0 {
			continue
		}

		// Split path into directory components
		dir := filepath.Dir(file.Path)
		parts := strings.Split(dir, string(filepath.Separator))

		// Navigate/create directory tree
		current := root
		currentPath := ""
		for _, part := range parts {
			if part == "." || part == "" {
				continue
			}
			if currentPath == "" {
				currentPath = part
			} else {
				currentPath = filepath.Join(currentPath, part)
			}

			if current.children[part] == nil {
				current.children[part] = &dirNode{
					name:     part,
					path:     currentPath,
					children: make(map[string]*dirNode),
				}
			}
			current = current.children[part]
		}

		// Add file to the appropriate directory
		current.files = append(current.files, file)
		lineCount := fileLineCount(file)
		current.lines += lineCount
	}

	// Recalculate line counts from bottom up
	calculateLines(root)

	return root
}

// calculateLines recursively calculates total lines for each node
func calculateLines(node *dirNode) int {
	total := 0

	// Count lines from files directly in this directory
	for _, file := range node.files {
		total += fileLineCount(file)
	}

	// Count lines from subdirectories
	for _, child := range node.children {
		total += calculateLines(child)
	}

	node.lines = total
	return total
}

// pruneToLimit prunes the tree to fit within the line limit
// Returns the pruned file list and a list of pruned directories
func pruneToLimit(root *dirNode, limit int) ([]FileIndex, []string) {
	if limit <= 0 || root.lines <= limit {
		// No pruning needed, collect all files
		return collectFiles(root), nil
	}

	var prunedDirs []string
	currentLines := root.lines

	// Keep pruning until we're under the limit
	for currentLines > limit {
		// Find the largest leaf node (directory with no subdirectories, or file)
		leaf, parent := findLargestLeaf(root)
		if leaf == nil {
			break // No more nodes to prune
		}

		// Remove the leaf
		if parent != nil {
			delete(parent.children, leaf.name)
		}

		// Track pruned directory
		if leaf.path != "" {
			prunedDirs = append(prunedDirs, leaf.path)
		}

		// Update line counts
		currentLines -= leaf.lines
		recalculateParentLines(root)
	}

	// If still over limit, prune files from directories
	if currentLines > limit {
		pruneFilesToLimit(root, &currentLines, limit)
	}

	return collectFiles(root), prunedDirs
}

// findLargestLeaf finds the leaf node (no children) with the most lines
func findLargestLeaf(root *dirNode) (*dirNode, *dirNode) {
	var largest *dirNode
	var largestParent *dirNode
	maxLines := 0

	var findLeaf func(node, parent *dirNode)
	findLeaf = func(node, parent *dirNode) {
		// If this node has no children, it's a leaf
		if len(node.children) == 0 && node != root {
			if node.lines > maxLines {
				maxLines = node.lines
				largest = node
				largestParent = parent
			}
			return
		}

		// Recurse into children
		for _, child := range node.children {
			findLeaf(child, node)
		}
	}

	findLeaf(root, nil)
	return largest, largestParent
}

// recalculateParentLines recalculates all line counts after pruning
func recalculateParentLines(root *dirNode) {
	calculateLines(root)
}

// pruneFilesToLimit removes files from directories to fit the limit
func pruneFilesToLimit(root *dirNode, currentLines *int, limit int) {
	// Collect all files with their line counts
	type fileEntry struct {
		file  *FileIndex
		node  *dirNode
		index int
		lines int
	}

	var entries []fileEntry
	var collectEntries func(node *dirNode)
	collectEntries = func(node *dirNode) {
		for i := range node.files {
			entries = append(entries, fileEntry{
				file:  &node.files[i],
				node:  node,
				index: i,
				lines: fileLineCount(node.files[i]),
			})
		}
		for _, child := range node.children {
			collectEntries(child)
		}
	}
	collectEntries(root)

	// Sort by line count descending (prune largest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].lines > entries[j].lines
	})

	// Remove files until under limit
	removed := make(map[*dirNode]map[int]bool)
	for _, entry := range entries {
		if *currentLines <= limit {
			break
		}
		if removed[entry.node] == nil {
			removed[entry.node] = make(map[int]bool)
		}
		removed[entry.node][entry.index] = true
		*currentLines -= entry.lines
	}

	// Actually remove the files
	for node, indices := range removed {
		var newFiles []FileIndex
		for i, file := range node.files {
			if !indices[i] {
				newFiles = append(newFiles, file)
			}
		}
		node.files = newFiles
	}
}

// collectFiles collects all files from the tree in sorted order
func collectFiles(root *dirNode) []FileIndex {
	var files []FileIndex

	var collect func(node *dirNode)
	collect = func(node *dirNode) {
		files = append(files, node.files...)
		// Sort children for deterministic order
		var childNames []string
		for name := range node.children {
			childNames = append(childNames, name)
		}
		sort.Strings(childNames)
		for _, name := range childNames {
			collect(node.children[name])
		}
	}

	collect(root)

	// Sort files by path
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	return files
}
