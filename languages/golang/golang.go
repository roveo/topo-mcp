//go:build lang_go || lang_all || (!lang_python && !lang_typescript && !lang_rust)

package golang

import (
	"context"
	"fmt"
	"strings"

	"github.com/roveo/topo-mcp/languages"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

func init() {
	languages.Register(&Language{})
}

// Language implements the Go language parser
type Language struct{}

func (g *Language) Name() string {
	return "go"
}

func (g *Language) Extensions() []string {
	return []string{".go"}
}

func (g *Language) TreeSitterLang() *sitter.Language {
	return golang.GetLanguage()
}

func (g *Language) Parse(content []byte) ([]string, []languages.Symbol, error) {
	parser := sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(golang.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse Go file: %w", err)
	}
	defer tree.Close()

	root := tree.RootNode()

	var imports []string
	var symbols []languages.Symbol

	// Walk top-level declarations
	for i := 0; i < int(root.NamedChildCount()); i++ {
		child := root.NamedChild(i)
		switch child.Type() {
		case "import_declaration":
			imports = append(imports, extractImports(child, content)...)
		case "function_declaration":
			symbols = append(symbols, extractFunction(child, content))
		case "method_declaration":
			symbols = append(symbols, extractMethod(child, content))
		case "type_declaration":
			symbols = append(symbols, extractTypes(child, content)...)
		case "const_declaration":
			symbols = append(symbols, extractConsts(child, content)...)
		case "var_declaration":
			symbols = append(symbols, extractVars(child, content)...)
		}
	}

	return imports, symbols, nil
}

// extractImports extracts import paths from an import_declaration
func extractImports(node *sitter.Node, content []byte) []string {
	var imports []string

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "import_spec" || child.Type() == "import_spec_list" {
			imports = append(imports, extractImportSpecs(child, content)...)
		}
	}

	return imports
}

func extractImportSpecs(node *sitter.Node, content []byte) []string {
	var imports []string

	if node.Type() == "import_spec" {
		pathNode := node.ChildByFieldName("path")
		if pathNode != nil {
			path := pathNode.Content(content)
			path = strings.Trim(path, `"`)
			imports = append(imports, path)
		}
	} else if node.Type() == "import_spec_list" {
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if child.Type() == "import_spec" {
				imports = append(imports, extractImportSpecs(child, content)...)
			}
		}
	}

	return imports
}

// extractFunction extracts a function declaration
func extractFunction(node *sitter.Node, content []byte) languages.Symbol {
	nameNode := node.ChildByFieldName("name")
	name := ""
	if nameNode != nil {
		name = nameNode.Content(content)
	}

	params := node.ChildByFieldName("parameters")
	result := node.ChildByFieldName("result")
	signature := formatSignature(params, result, content)

	doc := extractDoc(node, content)

	return &Function{
		name:      name,
		signature: signature,
		doc:       doc,
		loc:       languages.NodeRange(node),
	}
}

// extractMethod extracts a method declaration
func extractMethod(node *sitter.Node, content []byte) languages.Symbol {
	nameNode := node.ChildByFieldName("name")
	name := ""
	if nameNode != nil {
		name = nameNode.Content(content)
	}

	receiverNode := node.ChildByFieldName("receiver")
	receiver := formatReceiver(receiverNode, content)

	params := node.ChildByFieldName("parameters")
	result := node.ChildByFieldName("result")
	signature := formatSignature(params, result, content)

	doc := extractDoc(node, content)

	return &Method{
		name:      name,
		receiver:  receiver,
		signature: signature,
		doc:       doc,
		loc:       languages.NodeRange(node),
	}
}

// extractTypes extracts type declarations
func extractTypes(node *sitter.Node, content []byte) []languages.Symbol {
	var symbols []languages.Symbol

	doc := extractDoc(node, content)

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "type_spec" {
			nameNode := child.ChildByFieldName("name")
			name := ""
			if nameNode != nil {
				name = nameNode.Content(content)
			}

			typeNode := child.ChildByFieldName("type")
			typeKind := getTypeKind(typeNode, content)

			symbols = append(symbols, &Type{
				name:     name,
				typeKind: typeKind,
				doc:      doc,
				loc:      languages.NodeRange(child),
			})
		}
	}

	return symbols
}

// extractConsts extracts const declarations
func extractConsts(node *sitter.Node, content []byte) []languages.Symbol {
	var symbols []languages.Symbol

	doc := extractDoc(node, content)

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "const_spec" {
			names := extractSpecNames(child, content)
			for _, name := range names {
				symbols = append(symbols, &Const{
					name: name,
					doc:  doc,
					loc:  languages.NodeRange(child),
				})
			}
		}
	}

	return symbols
}

// extractVars extracts var declarations
func extractVars(node *sitter.Node, content []byte) []languages.Symbol {
	var symbols []languages.Symbol

	doc := extractDoc(node, content)

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "var_spec" {
			names := extractSpecNames(child, content)
			for _, name := range names {
				symbols = append(symbols, &Var{
					name: name,
					doc:  doc,
					loc:  languages.NodeRange(child),
				})
			}
		}
	}

	return symbols
}

// extractSpecNames extracts identifier names from a const_spec or var_spec
func extractSpecNames(node *sitter.Node, content []byte) []string {
	var names []string

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "identifier" {
			names = append(names, child.Content(content))
		}
	}

	return names
}

// formatSignature formats function parameters and return types
func formatSignature(params, result *sitter.Node, content []byte) string {
	var sb strings.Builder

	sb.WriteString("(")
	if params != nil {
		var paramTypes []string
		for i := 0; i < int(params.NamedChildCount()); i++ {
			child := params.NamedChild(i)
			if child.Type() == "parameter_declaration" {
				typeNode := child.ChildByFieldName("type")
				if typeNode != nil {
					typeStr := typeNode.Content(content)
					nameCount := 0
					for j := 0; j < int(child.NamedChildCount()); j++ {
						if child.NamedChild(j).Type() == "identifier" {
							nameCount++
						}
					}
					if nameCount == 0 {
						nameCount = 1
					}
					for k := 0; k < nameCount; k++ {
						paramTypes = append(paramTypes, typeStr)
					}
				}
			}
		}
		sb.WriteString(strings.Join(paramTypes, ", "))
	}
	sb.WriteString(")")

	if result != nil {
		resultStr := formatResult(result, content)
		if resultStr != "" {
			sb.WriteString(" ")
			sb.WriteString(resultStr)
		}
	}

	return sb.String()
}

// formatResult formats the return type(s)
func formatResult(node *sitter.Node, content []byte) string {
	if node == nil {
		return ""
	}

	if node.Type() != "parameter_list" {
		return node.Content(content)
	}

	var types []string
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "parameter_declaration" {
			typeNode := child.ChildByFieldName("type")
			if typeNode != nil {
				types = append(types, typeNode.Content(content))
			}
		}
	}

	if len(types) == 1 {
		return types[0]
	}
	return "(" + strings.Join(types, ", ") + ")"
}

// formatReceiver formats the method receiver
func formatReceiver(node *sitter.Node, content []byte) string {
	if node == nil {
		return ""
	}

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "parameter_declaration" {
			typeNode := child.ChildByFieldName("type")
			if typeNode != nil {
				return typeNode.Content(content)
			}
		}
	}

	return ""
}

// getTypeKind determines if a type is struct, interface, or an alias
func getTypeKind(node *sitter.Node, content []byte) string {
	if node == nil {
		return ""
	}

	switch node.Type() {
	case "struct_type":
		return "struct"
	case "interface_type":
		return "interface"
	default:
		return node.Content(content)
	}
}

// extractDoc extracts the first line of the doc comment
func extractDoc(node *sitter.Node, content []byte) string {
	prev := node.PrevNamedSibling()
	if prev == nil || prev.Type() != "comment" {
		return ""
	}

	commentEndLine := prev.EndPoint().Row
	declStartLine := node.StartPoint().Row
	if declStartLine-commentEndLine > 1 {
		return ""
	}

	text := prev.Content(content)
	text = strings.TrimPrefix(text, "//")
	text = strings.TrimPrefix(text, "/*")
	text = strings.TrimSuffix(text, "*/")
	text = strings.TrimSpace(text)

	for line := range strings.SplitSeq(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}

	return ""
}
