package python

import (
	"context"
	"fmt"
	"strings"

	"github.com/roveo/topo-mcp/languages"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
)

func init() {
	languages.Register(&Language{})
}

// Language implements the Python language parser
type Language struct{}

func (p *Language) Name() string {
	return "python"
}

func (p *Language) Extensions() []string {
	return []string{".py"}
}

func (p *Language) Parse(content []byte) ([]string, []languages.Symbol, error) {
	parser := sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(python.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse Python file: %w", err)
	}
	defer tree.Close()

	root := tree.RootNode()

	var imports []string
	var symbols []languages.Symbol

	for i := 0; i < int(root.NamedChildCount()); i++ {
		child := root.NamedChild(i)
		switch child.Type() {
		case "import_statement":
			imports = append(imports, extractImport(child, content)...)
		case "import_from_statement":
			imports = append(imports, extractFromImport(child, content)...)
		case "function_definition":
			symbols = append(symbols, extractFunction(child, content))
		case "class_definition":
			symbols = append(symbols, extractClass(child, content))
		case "decorated_definition":
			symbols = append(symbols, extractDecorated(child, content)...)
		case "expression_statement":
			if assign := extractAssignment(child, content); assign != nil {
				symbols = append(symbols, assign...)
			}
		}
	}

	return imports, symbols, nil
}

func extractImport(node *sitter.Node, content []byte) []string {
	var imports []string

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "dotted_name" || child.Type() == "aliased_import" {
			name := extractDottedName(child, content)
			if name != "" {
				imports = append(imports, name)
			}
		}
	}

	return imports
}

func extractFromImport(node *sitter.Node, content []byte) []string {
	var imports []string

	moduleName := ""
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "dotted_name" || child.Type() == "relative_import" {
			moduleName = child.Content(content)
			break
		}
	}

	if moduleName != "" {
		imports = append(imports, moduleName)
	}

	return imports
}

func extractDottedName(node *sitter.Node, content []byte) string {
	if node.Type() == "aliased_import" {
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil {
			return nameNode.Content(content)
		}
	}
	return node.Content(content)
}

func extractFunction(node *sitter.Node, content []byte) languages.Symbol {
	nameNode := node.ChildByFieldName("name")
	name := ""
	if nameNode != nil {
		name = nameNode.Content(content)
	}

	params := node.ChildByFieldName("parameters")
	returnType := node.ChildByFieldName("return_type")
	signature := formatSignature(params, returnType, content)

	doc := extractDocstring(node, content)

	return &Function{
		name:      name,
		signature: signature,
		doc:       doc,
		loc:       languages.NodeRange(node),
	}
}

func extractClass(node *sitter.Node, content []byte) languages.Symbol {
	nameNode := node.ChildByFieldName("name")
	name := ""
	if nameNode != nil {
		name = nameNode.Content(content)
	}

	var bases []string
	superclass := node.ChildByFieldName("superclasses")
	if superclass != nil {
		bases = extractBases(superclass, content)
	}

	doc := extractDocstring(node, content)

	return &Class{
		name:  name,
		bases: bases,
		doc:   doc,
		loc:   languages.NodeRange(node),
	}
}

func extractDecorated(node *sitter.Node, content []byte) []languages.Symbol {
	var symbols []languages.Symbol
	var decorators []string

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		switch child.Type() {
		case "decorator":
			dec := extractDecorator(child, content)
			if dec != "" {
				decorators = append(decorators, dec)
			}
		case "function_definition":
			sym := extractFunction(child, content)
			if fn, ok := sym.(*Function); ok {
				fn.decorators = decorators
			}
			symbols = append(symbols, sym)
		case "class_definition":
			sym := extractClass(child, content)
			if cls, ok := sym.(*Class); ok {
				cls.decorators = decorators
			}
			symbols = append(symbols, sym)
		}
	}

	return symbols
}

func extractDecorator(node *sitter.Node, content []byte) string {
	text := node.Content(content)
	text = strings.TrimPrefix(text, "@")
	if idx := strings.Index(text, "("); idx != -1 {
		text = text[:idx]
	}
	return strings.TrimSpace(text)
}

func extractBases(node *sitter.Node, content []byte) []string {
	var bases []string

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() != "keyword_argument" {
			bases = append(bases, child.Content(content))
		}
	}

	return bases
}

func extractAssignment(node *sitter.Node, content []byte) []languages.Symbol {
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "assignment" {
			return extractAssignmentTargets(child, content)
		}
	}
	return nil
}

func extractAssignmentTargets(node *sitter.Node, content []byte) []languages.Symbol {
	var symbols []languages.Symbol

	left := node.ChildByFieldName("left")
	if left == nil {
		return nil
	}

	if left.Type() == "identifier" {
		name := left.Content(content)
		if !strings.HasPrefix(name, "_") {
			symbols = append(symbols, &Variable{
				name: name,
				loc:  languages.NodeRange(node),
			})
		}
	}

	return symbols
}

func formatSignature(params, returnType *sitter.Node, content []byte) string {
	var sb strings.Builder

	if params != nil {
		sb.WriteString(params.Content(content))
	} else {
		sb.WriteString("()")
	}

	if returnType != nil {
		sb.WriteString(" -> ")
		sb.WriteString(returnType.Content(content))
	}

	return sb.String()
}

func extractDocstring(node *sitter.Node, content []byte) string {
	body := node.ChildByFieldName("body")
	if body == nil {
		return ""
	}

	if body.NamedChildCount() > 0 {
		first := body.NamedChild(0)
		if first.Type() == "expression_statement" {
			if first.NamedChildCount() > 0 {
				expr := first.NamedChild(0)
				if expr.Type() == "string" {
					docstring := expr.Content(content)
					return cleanDocstring(docstring)
				}
			}
		}
	}

	return ""
}

func cleanDocstring(s string) string {
	s = strings.Trim(s, `"'`)
	s = strings.TrimPrefix(s, `""`)
	s = strings.TrimSuffix(s, `""`)
	s = strings.TrimPrefix(s, `''`)
	s = strings.TrimSuffix(s, `''`)

	for line := range strings.SplitSeq(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}

	return ""
}
