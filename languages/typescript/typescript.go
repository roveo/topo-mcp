package typescript

import (
	"context"
	"fmt"
	"strings"

	"github.com/roveo/topo-mcp/languages"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

func init() {
	languages.Register(&TSLanguage{})
	languages.Register(&TSXLanguage{})
	languages.Register(&JSLanguage{})
	languages.Register(&JSXLanguage{})
}

// TSLanguage implements TypeScript (.ts) parsing
type TSLanguage struct{}

func (t *TSLanguage) Name() string         { return "typescript" }
func (t *TSLanguage) Extensions() []string { return []string{".ts"} }
func (t *TSLanguage) Parse(content []byte) ([]string, []languages.Symbol, error) {
	return parse(content, typescript.GetLanguage(), "typescript")
}

// TSXLanguage implements TSX (.tsx) parsing
type TSXLanguage struct{}

func (t *TSXLanguage) Name() string         { return "tsx" }
func (t *TSXLanguage) Extensions() []string { return []string{".tsx"} }
func (t *TSXLanguage) Parse(content []byte) ([]string, []languages.Symbol, error) {
	return parse(content, tsx.GetLanguage(), "tsx")
}

// JSLanguage implements JavaScript (.js) parsing
type JSLanguage struct{}

func (j *JSLanguage) Name() string         { return "javascript" }
func (j *JSLanguage) Extensions() []string { return []string{".js", ".mjs", ".cjs"} }
func (j *JSLanguage) Parse(content []byte) ([]string, []languages.Symbol, error) {
	return parse(content, javascript.GetLanguage(), "javascript")
}

// JSXLanguage implements JSX (.jsx) parsing
type JSXLanguage struct{}

func (j *JSXLanguage) Name() string         { return "jsx" }
func (j *JSXLanguage) Extensions() []string { return []string{".jsx"} }
func (j *JSXLanguage) Parse(content []byte) ([]string, []languages.Symbol, error) {
	return parse(content, javascript.GetLanguage(), "jsx")
}

func parse(content []byte, lang *sitter.Language, langName string) ([]string, []languages.Symbol, error) {
	parser := sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(lang)

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse %s file: %w", langName, err)
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
		case "function_declaration":
			symbols = append(symbols, extractFunction(child, content))
		case "class_declaration":
			symbols = append(symbols, extractClass(child, content))
		case "interface_declaration":
			symbols = append(symbols, extractInterface(child, content))
		case "type_alias_declaration":
			symbols = append(symbols, extractTypeAlias(child, content))
		case "enum_declaration":
			symbols = append(symbols, extractEnum(child, content))
		case "lexical_declaration", "variable_declaration":
			symbols = append(symbols, extractVariables(child, content)...)
		case "export_statement":
			syms, imps := extractExport(child, content)
			symbols = append(symbols, syms...)
			imports = append(imports, imps...)
		}
	}

	return imports, symbols, nil
}

func extractImport(node *sitter.Node, content []byte) []string {
	var imports []string

	source := node.ChildByFieldName("source")
	if source != nil {
		path := source.Content(content)
		path = strings.Trim(path, `"'`)
		imports = append(imports, path)
	}

	return imports
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

	isAsync := hasChildOfType(node, "async")
	doc := extractDoc(node, content)

	return &Function{
		name:      name,
		signature: signature,
		isAsync:   isAsync,
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

	var extends string
	var implements []string

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "class_heritage" {
			extends, implements = extractHeritage(child, content)
		}
	}

	doc := extractDoc(node, content)

	return &Class{
		name:       name,
		extends:    extends,
		implements: implements,
		doc:        doc,
		loc:        languages.NodeRange(node),
	}
}

func extractHeritage(node *sitter.Node, content []byte) (string, []string) {
	var extends string
	var implements []string

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		switch child.Type() {
		case "extends_clause":
			if child.NamedChildCount() > 0 {
				extends = child.NamedChild(0).Content(content)
			}
		case "implements_clause":
			for j := 0; j < int(child.NamedChildCount()); j++ {
				impl := child.NamedChild(j)
				implements = append(implements, impl.Content(content))
			}
		}
	}

	return extends, implements
}

func extractInterface(node *sitter.Node, content []byte) languages.Symbol {
	nameNode := node.ChildByFieldName("name")
	name := ""
	if nameNode != nil {
		name = nameNode.Content(content)
	}

	doc := extractDoc(node, content)

	return &Interface{
		name: name,
		doc:  doc,
		loc:  languages.NodeRange(node),
	}
}

func extractTypeAlias(node *sitter.Node, content []byte) languages.Symbol {
	nameNode := node.ChildByFieldName("name")
	name := ""
	if nameNode != nil {
		name = nameNode.Content(content)
	}

	doc := extractDoc(node, content)

	return &TypeAlias{
		name: name,
		doc:  doc,
		loc:  languages.NodeRange(node),
	}
}

func extractEnum(node *sitter.Node, content []byte) languages.Symbol {
	nameNode := node.ChildByFieldName("name")
	name := ""
	if nameNode != nil {
		name = nameNode.Content(content)
	}

	doc := extractDoc(node, content)

	return &Enum{
		name: name,
		doc:  doc,
		loc:  languages.NodeRange(node),
	}
}

func extractVariables(node *sitter.Node, content []byte) []languages.Symbol {
	var symbols []languages.Symbol

	kind := "var"
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		text := child.Content(content)
		if text == "const" {
			kind = "const"
			break
		} else if text == "let" {
			kind = "let"
			break
		}
	}

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "variable_declarator" {
			nameNode := child.ChildByFieldName("name")
			if nameNode != nil {
				name := nameNode.Content(content)
				symbols = append(symbols, &Variable{
					name: name,
					kind: kind,
					loc:  languages.NodeRange(child),
				})
			}
		}
	}

	return symbols
}

func extractExport(node *sitter.Node, content []byte) ([]languages.Symbol, []string) {
	var symbols []languages.Symbol
	var imports []string

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		switch child.Type() {
		case "function_declaration":
			symbols = append(symbols, extractFunction(child, content))
		case "class_declaration":
			symbols = append(symbols, extractClass(child, content))
		case "interface_declaration":
			symbols = append(symbols, extractInterface(child, content))
		case "type_alias_declaration":
			symbols = append(symbols, extractTypeAlias(child, content))
		case "enum_declaration":
			symbols = append(symbols, extractEnum(child, content))
		case "lexical_declaration", "variable_declaration":
			symbols = append(symbols, extractVariables(child, content)...)
		}
	}

	return symbols, imports
}

func formatSignature(params, returnType *sitter.Node, content []byte) string {
	var sb strings.Builder

	if params != nil {
		sb.WriteString(params.Content(content))
	} else {
		sb.WriteString("()")
	}

	if returnType != nil {
		retStr := returnType.Content(content)
		// The return_type node sometimes includes the colon, sometimes not
		if !strings.HasPrefix(retStr, ":") {
			sb.WriteString(": ")
		}
		sb.WriteString(retStr)
	}

	return sb.String()
}

func hasChildOfType(node *sitter.Node, typeName string) bool {
	for i := 0; i < int(node.ChildCount()); i++ {
		if node.Child(i).Type() == typeName {
			return true
		}
	}
	return false
}

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

	if strings.HasPrefix(text, "/**") {
		text = strings.TrimPrefix(text, "/**")
		text = strings.TrimSuffix(text, "*/")
	} else if strings.HasPrefix(text, "/*") {
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
	} else if strings.HasPrefix(text, "//") {
		text = strings.TrimPrefix(text, "//")
	}

	for line := range strings.SplitSeq(text, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "@") {
			return line
		}
	}

	return ""
}
