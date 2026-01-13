// Package gitignore provides functionality to parse and match .gitignore patterns.
package gitignore

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Matcher holds compiled gitignore patterns for a directory tree.
type Matcher struct {
	root     string
	patterns []pattern
}

// pattern represents a single gitignore pattern with its context.
type pattern struct {
	pattern  string // The original pattern (cleaned)
	negation bool   // Pattern starts with !
	dirOnly  bool   // Pattern ends with /
	anchored bool   // Pattern contains / (except trailing)
	baseDir  string // Directory where the .gitignore was found (relative to root)
}

// New creates a new Matcher for the given root directory.
// It recursively loads all .gitignore files in the directory tree.
func New(root string) (*Matcher, error) {
	m := &Matcher{root: root}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip inaccessible paths
		}

		// Skip hidden directories (but not .gitignore files)
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		if info.Name() == ".gitignore" {
			relDir, _ := filepath.Rel(root, filepath.Dir(path))
			if relDir == "." {
				relDir = ""
			}
			if err := m.loadFile(path, relDir); err != nil {
				return nil // Skip unreadable .gitignore files
			}
		}

		return nil
	})

	return m, err
}

// loadFile parses a .gitignore file and adds its patterns.
func (m *Matcher) loadFile(path string, baseDir string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if p := parseLine(line, baseDir); p != nil {
			m.patterns = append(m.patterns, *p)
		}
	}

	return scanner.Err()
}

// parseLine parses a single line from a .gitignore file.
// Returns nil for empty lines and comments.
func parseLine(line string, baseDir string) *pattern {
	// Trim trailing spaces (unless escaped)
	for len(line) > 0 && line[len(line)-1] == ' ' {
		if len(line) >= 2 && line[len(line)-2] == '\\' {
			line = line[:len(line)-2] + " "
			break
		}
		line = line[:len(line)-1]
	}

	// Skip empty lines and comments
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}

	p := &pattern{baseDir: baseDir}

	// Check for negation
	if strings.HasPrefix(line, "!") {
		p.negation = true
		line = line[1:]
	}

	// Check for directory-only match
	if strings.HasSuffix(line, "/") {
		p.dirOnly = true
		line = strings.TrimSuffix(line, "/")
	}

	// Check if pattern is anchored (contains / except at end)
	// A leading / also anchors the pattern
	if strings.HasPrefix(line, "/") {
		p.anchored = true
		line = line[1:]
	} else if strings.Contains(line, "/") {
		p.anchored = true
	}

	p.pattern = line
	return p
}

// Match checks if a path should be ignored.
// The path should be relative to the Matcher's root directory.
// isDir should be true if the path is a directory.
// Match checks if a path should be ignored.
// The path should be relative to the Matcher's root directory.
// isDir should be true if the path is a directory.
func (m *Matcher) Match(path string, isDir bool) bool {
	if m == nil || len(m.patterns) == 0 {
		return false
	}

	// Normalize path separators
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, "./")

	// Check if any parent directory is ignored
	// This is needed because gitignore patterns like "build/" should also
	// ignore all files under build/
	parts := strings.Split(path, "/")
	for i := 1; i < len(parts); i++ {
		parentPath := strings.Join(parts[:i], "/")
		if m.matchPath(parentPath, true) {
			return true
		}
	}

	return m.matchPath(path, isDir)
}

// matchPath checks if a specific path matches the gitignore patterns.
func (m *Matcher) matchPath(path string, isDir bool) bool {
	ignored := false

	for _, p := range m.patterns {
		if p.matches(path, isDir) {
			ignored = !p.negation
		}
	}

	return ignored
}

// matches checks if a single pattern matches the given path.
func (p *pattern) matches(path string, isDir bool) bool {
	// Directory-only patterns don't match files
	if p.dirOnly && !isDir {
		return false
	}

	// If the pattern is from a subdirectory, the path must be under that directory
	if p.baseDir != "" {
		if !strings.HasPrefix(path, p.baseDir+"/") {
			return false
		}
		// Make path relative to the pattern's base directory
		path = strings.TrimPrefix(path, p.baseDir+"/")
	}

	// Anchored patterns must match from the start
	if p.anchored {
		return matchGlob(p.pattern, path)
	}

	// Non-anchored patterns can match at any directory level
	// Try matching against the full path first
	if matchGlob(p.pattern, path) {
		return true
	}

	// Try matching against each path component
	parts := strings.Split(path, "/")
	for i := range parts {
		subpath := strings.Join(parts[i:], "/")
		if matchGlob(p.pattern, subpath) {
			return true
		}
	}

	return false
}

// matchGlob performs glob-style pattern matching.
// Supports *, **, and ? wildcards.
func matchGlob(pattern, name string) bool {
	// Handle ** (matches any number of directories)
	if strings.Contains(pattern, "**") {
		return matchDoublestar(pattern, name)
	}

	return matchSimpleGlob(pattern, name)
}

// matchDoublestar handles patterns containing **.
func matchDoublestar(pattern, name string) bool {
	// Split pattern by **
	parts := strings.Split(pattern, "**")

	if len(parts) == 2 {
		prefix := parts[0]
		suffix := strings.TrimPrefix(parts[1], "/")

		// Prefix must match the start
		if prefix != "" {
			prefix = strings.TrimSuffix(prefix, "/")
			if !strings.HasPrefix(name, prefix) {
				return false
			}
			name = strings.TrimPrefix(name, prefix)
			name = strings.TrimPrefix(name, "/")
		}

		// Suffix must match the end (or any subpath for directories)
		if suffix == "" {
			return true
		}

		// Try matching suffix at each level
		if matchSimpleGlob(suffix, name) {
			return true
		}

		pathParts := strings.Split(name, "/")
		for i := range pathParts {
			subpath := strings.Join(pathParts[i:], "/")
			if matchSimpleGlob(suffix, subpath) {
				return true
			}
		}

		return false
	}

	// Multiple ** in pattern - use recursive approach
	return matchDoublestarRecursive(pattern, name)
}

// matchDoublestarRecursive handles complex patterns with multiple **.
func matchDoublestarRecursive(pattern, name string) bool {
	idx := strings.Index(pattern, "**")
	if idx == -1 {
		return matchSimpleGlob(pattern, name)
	}

	prefix := pattern[:idx]
	suffix := pattern[idx+2:]
	suffix = strings.TrimPrefix(suffix, "/")

	// The prefix must match
	if prefix != "" {
		prefix = strings.TrimSuffix(prefix, "/")
		if !hasPrefix(name, prefix) {
			return false
		}
		name = strings.TrimPrefix(name, prefix)
		name = strings.TrimPrefix(name, "/")
	}

	// Try matching the rest at each level
	if matchDoublestarRecursive(suffix, name) {
		return true
	}

	parts := strings.Split(name, "/")
	for i := 1; i <= len(parts); i++ {
		subpath := strings.Join(parts[i:], "/")
		if matchDoublestarRecursive(suffix, subpath) {
			return true
		}
	}

	return false
}

// hasPrefix checks if name starts with prefix using glob matching.
func hasPrefix(name, prefix string) bool {
	if len(name) < len(prefix) {
		return matchSimpleGlob(prefix, name)
	}
	return matchSimpleGlob(prefix, name[:len(prefix)]) ||
		(len(name) > len(prefix) && name[len(prefix)] == '/' && matchSimpleGlob(prefix, name[:len(prefix)]))
}

// matchSimpleGlob matches patterns with * and ? but not **.
func matchSimpleGlob(pattern, name string) bool {
	px, nx := 0, 0
	starPx, starNx := -1, -1

	for nx < len(name) {
		if px < len(pattern) {
			switch pattern[px] {
			case '*':
				// Remember this position for backtracking
				starPx = px
				starNx = nx
				px++
				continue
			case '?':
				// Match any single character (except /)
				if name[nx] == '/' {
					goto backtrack
				}
				px++
				nx++
				continue
			default:
				if pattern[px] == name[nx] {
					px++
					nx++
					continue
				}
			}
		}

	backtrack:
		// Try to match more with the last *
		if starPx >= 0 && starNx < len(name) && name[starNx] != '/' {
			starNx++
			px = starPx + 1
			nx = starNx
			continue
		}

		return false
	}

	// Skip trailing *s in pattern
	for px < len(pattern) && pattern[px] == '*' {
		px++
	}

	return px == len(pattern)
}
