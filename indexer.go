package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type Symbol struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Receiver  string `json:"receiver,omitempty"`
	TypeKind  string `json:"type_kind,omitempty"` // For types: struct, interface, or the underlying type
	Signature string `json:"signature,omitempty"` // For functions: (params) returns
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	DocHead   string `json:"doc_head,omitempty"`
}

type FileIndex struct {
	Path    string   `json:"path"`
	Imports []string `json:"imports"`
	Symbols []Symbol `json:"symbols"`
}

type PackageIndex struct {
	Path  string      `json:"path"`
	Dir   string      `json:"dir"`
	Files []FileIndex `json:"files"`
}

// FormatOptions controls how the index is formatted
type FormatOptions struct {
	SkipPatterns []string // Path prefixes to skip by default
	Filter       string   // If set, only show files matching this prefix (overrides skip)
}

func firstLineOfComment(cg *ast.CommentGroup) string {
	if cg == nil || len(cg.List) == 0 {
		return ""
	}
	text := cg.Text()
	for ln := range strings.SplitSeq(text, "\n") {
		s := strings.TrimSpace(strings.TrimPrefix(ln, "//"))
		if s != "" {
			return s
		}
	}
	return ""
}

// formatExpr formats an AST expression as a string (for types)
func formatExpr(expr ast.Expr) string {
	if expr == nil {
		return ""
	}
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + formatExpr(t.X)
	case *ast.SelectorExpr:
		return formatExpr(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + formatExpr(t.Elt)
		}
		return "[...]" + formatExpr(t.Elt)
	case *ast.MapType:
		return "map[" + formatExpr(t.Key) + "]" + formatExpr(t.Value)
	case *ast.ChanType:
		switch t.Dir {
		case ast.SEND:
			return "chan<- " + formatExpr(t.Value)
		case ast.RECV:
			return "<-chan " + formatExpr(t.Value)
		default:
			return "chan " + formatExpr(t.Value)
		}
	case *ast.FuncType:
		return "func" + formatFuncSignature(t)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	case *ast.Ellipsis:
		return "..." + formatExpr(t.Elt)
	default:
		return ""
	}
}

// formatFuncSignature formats function parameters and return types
func formatFuncSignature(ft *ast.FuncType) string {
	var sb strings.Builder

	// Format parameters
	sb.WriteString("(")
	if ft.Params != nil {
		var params []string
		for _, field := range ft.Params.List {
			typeStr := formatExpr(field.Type)
			// If multiple names share the same type, we just output the type once per name
			if len(field.Names) == 0 {
				params = append(params, typeStr)
			} else {
				for range field.Names {
					params = append(params, typeStr)
				}
			}
		}
		sb.WriteString(strings.Join(params, ", "))
	}
	sb.WriteString(")")

	// Format return types
	if ft.Results != nil && len(ft.Results.List) > 0 {
		var results []string
		for _, field := range ft.Results.List {
			typeStr := formatExpr(field.Type)
			if len(field.Names) == 0 {
				results = append(results, typeStr)
			} else {
				for range field.Names {
					results = append(results, typeStr)
				}
			}
		}
		if len(results) == 1 {
			sb.WriteString(" " + results[0])
		} else {
			sb.WriteString(" (" + strings.Join(results, ", ") + ")")
		}
	}

	return sb.String()
}

// getTypeKind returns "struct", "interface", or the underlying type expression
func getTypeKind(typeExpr ast.Expr) string {
	switch t := typeExpr.(type) {
	case *ast.StructType:
		return "struct"
	case *ast.InterfaceType:
		return "interface"
	default:
		// For type aliases like "type MyInt int", return the underlying type
		return formatExpr(t)
	}
}

func symbolKind(d ast.Decl) string {
	switch d := d.(type) {
	case *ast.FuncDecl:
		return "func"
	case *ast.GenDecl:
		switch d.Tok {
		case token.TYPE:
			return "type"
		case token.CONST:
			return "const"
		case token.VAR:
			return "var"
		}
	}
	return "unknown"
}

func collectSymbols(fset *token.FileSet, f *ast.File) []Symbol {
	var out []Symbol

	for _, d := range f.Decls {
		switch decl := d.(type) {

		case *ast.FuncDecl:
			start := fset.Position(decl.Pos()).Line
			end := fset.Position(decl.End()).Line
			recv := ""
			if decl.Recv != nil && len(decl.Recv.List) > 0 {
				switch t := decl.Recv.List[0].Type.(type) {
				case *ast.Ident:
					recv = t.Name
				case *ast.StarExpr:
					if id, ok := t.X.(*ast.Ident); ok {
						recv = "*" + id.Name
					}
				}
			}
			out = append(out, Symbol{
				Name:      decl.Name.Name,
				Kind:      "func",
				Receiver:  recv,
				Signature: formatFuncSignature(decl.Type),
				StartLine: start,
				EndLine:   end,
				DocHead:   firstLineOfComment(decl.Doc),
			})

		case *ast.GenDecl:
			k := symbolKind(decl)
			for _, spec := range decl.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					start := fset.Position(s.Pos()).Line
					end := fset.Position(s.End()).Line
					out = append(out, Symbol{
						Name:      s.Name.Name,
						Kind:      k,
						TypeKind:  getTypeKind(s.Type),
						StartLine: start,
						EndLine:   end,
						DocHead:   firstLineOfComment(decl.Doc),
					})
				case *ast.ValueSpec:
					for _, name := range s.Names {
						start := fset.Position(name.Pos()).Line
						end := fset.Position(name.End()).Line
						out = append(out, Symbol{
							Name:      name.Name,
							Kind:      k,
							StartLine: start,
							EndLine:   end,
							DocHead:   firstLineOfComment(decl.Doc),
						})
					}
				}
			}
		}
	}
	return out
}

// indexDirectory walks the directory and indexes all Go files
func indexDirectory(dir string) ([]FileIndex, error) {
	var results []FileIndex

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and vendor
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			// Skip files that can't be parsed
			return nil
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			relPath = path
		}

		// Collect imports
		var imports []string
		for _, imp := range f.Imports {
			imports = append(imports, strings.Trim(imp.Path.Value, `"`))
		}

		symbols := collectSymbols(fset, f)

		results = append(results, FileIndex{
			Path:    relPath,
			Imports: imports,
			Symbols: symbols,
		})

		return nil
	})

	return results, err
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

// formatCompact formats the index in a compact human-readable format
func formatCompact(files []FileIndex, opts FormatOptions) string {
	var sb strings.Builder

	for _, file := range files {
		// Check if file matches filter (if specified)
		if opts.Filter != "" {
			if !matchesFilter(file.Path, opts.Filter) {
				continue // Don't show at all if filter is set and doesn't match
			}
		}

		// Check if file is skipped by default (only when no filter is set)
		if opts.Filter == "" && isSkipped(file.Path, opts.SkipPatterns) {
			sb.WriteString(fmt.Sprintf("## %s\n", file.Path))
			sb.WriteString("  (skipped by default - use filter parameter to index this path explicitly)\n\n")
			continue
		}

		if len(file.Symbols) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("## %s\n", file.Path))

		for _, sym := range file.Symbols {
			var line string
			switch sym.Kind {
			case "func":
				if sym.Receiver != "" {
					line = fmt.Sprintf("  (%s) %s%s [%d-%d]", sym.Receiver, sym.Name, sym.Signature, sym.StartLine, sym.EndLine)
				} else {
					line = fmt.Sprintf("  %s%s [%d-%d]", sym.Name, sym.Signature, sym.StartLine, sym.EndLine)
				}
			case "type":
				line = fmt.Sprintf("  type %s %s [%d-%d]", sym.Name, sym.TypeKind, sym.StartLine, sym.EndLine)
			case "const":
				line = fmt.Sprintf("  const %s [%d]", sym.Name, sym.StartLine)
			case "var":
				line = fmt.Sprintf("  var %s [%d]", sym.Name, sym.StartLine)
			default:
				line = fmt.Sprintf("  %s %s [%d-%d]", sym.Kind, sym.Name, sym.StartLine, sym.EndLine)
			}
			// Add docstring for types and functions
			if sym.DocHead != "" && (sym.Kind == "type" || sym.Kind == "func") {
				line += " // " + sym.DocHead
			}
			sb.WriteString(line + "\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
