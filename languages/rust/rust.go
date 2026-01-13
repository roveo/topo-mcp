//go:build lang_rust || lang_all

package rust

import (
	"context"
	"fmt"
	"strings"

	"github.com/roveo/topo-mcp/languages"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/rust"
)

func init() {
	languages.Register(&Language{})
}

// Language implements the Rust language parser
type Language struct{}

func (r *Language) Name() string         { return "rust" }
func (r *Language) Extensions() []string { return []string{".rs"} }

func (r *Language) Parse(content []byte) ([]string, []languages.Symbol, error) {
	parser := sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(rust.GetLanguage())

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse Rust file: %w", err)
	}
	defer tree.Close()

	root := tree.RootNode()

	var imports []string
	var symbols []languages.Symbol

	for i := 0; i < int(root.NamedChildCount()); i++ {
		child := root.NamedChild(i)
		switch child.Type() {
		case "use_declaration":
			imports = append(imports, extractUse(child, content)...)
		case "function_item":
			symbols = append(symbols, extractFunction(child, content))
		case "struct_item":
			symbols = append(symbols, extractStruct(child, content))
		case "enum_item":
			symbols = append(symbols, extractEnum(child, content))
		case "trait_item":
			symbols = append(symbols, extractTrait(child, content))
		case "impl_item":
			symbols = append(symbols, extractImpl(child, content)...)
		case "const_item":
			symbols = append(symbols, extractConst(child, content))
		case "static_item":
			symbols = append(symbols, extractStatic(child, content))
		case "type_item":
			symbols = append(symbols, extractTypeAlias(child, content))
		case "mod_item":
			symbols = append(symbols, extractMod(child, content))
		}
	}

	return imports, symbols, nil
}

func extractUse(node *sitter.Node, content []byte) []string {
	var imports []string

	// Get the use path
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "use_tree" || child.Type() == "scoped_identifier" {
			path := child.Content(content)
			imports = append(imports, path)
		}
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

	vis := extractVisibility(node, content)
	doc := extractDoc(node, content)

	return &Function{
		name:       name,
		signature:  signature,
		visibility: vis,
		doc:        doc,
		loc:        languages.NodeRange(node),
	}
}

func extractStruct(node *sitter.Node, content []byte) languages.Symbol {
	nameNode := node.ChildByFieldName("name")
	name := ""
	if nameNode != nil {
		name = nameNode.Content(content)
	}

	vis := extractVisibility(node, content)
	doc := extractDoc(node, content)

	return &Struct{
		name:       name,
		visibility: vis,
		doc:        doc,
		loc:        languages.NodeRange(node),
	}
}

func extractEnum(node *sitter.Node, content []byte) languages.Symbol {
	nameNode := node.ChildByFieldName("name")
	name := ""
	if nameNode != nil {
		name = nameNode.Content(content)
	}

	vis := extractVisibility(node, content)
	doc := extractDoc(node, content)

	return &Enum{
		name:       name,
		visibility: vis,
		doc:        doc,
		loc:        languages.NodeRange(node),
	}
}

func extractTrait(node *sitter.Node, content []byte) languages.Symbol {
	nameNode := node.ChildByFieldName("name")
	name := ""
	if nameNode != nil {
		name = nameNode.Content(content)
	}

	vis := extractVisibility(node, content)
	doc := extractDoc(node, content)

	return &Trait{
		name:       name,
		visibility: vis,
		doc:        doc,
		loc:        languages.NodeRange(node),
	}
}

func extractImpl(node *sitter.Node, content []byte) []languages.Symbol {
	var symbols []languages.Symbol

	// Get the type being implemented
	typeName := ""
	traitName := ""

	typeNode := node.ChildByFieldName("type")
	if typeNode != nil {
		typeName = typeNode.Content(content)
	}

	traitNode := node.ChildByFieldName("trait")
	if traitNode != nil {
		traitName = traitNode.Content(content)
	}

	// Extract methods from the impl body
	body := node.ChildByFieldName("body")
	if body != nil {
		for i := 0; i < int(body.NamedChildCount()); i++ {
			child := body.NamedChild(i)
			if child.Type() == "function_item" {
				sym := extractFunction(child, content)
				if fn, ok := sym.(*Function); ok {
					fn.receiver = typeName
					fn.traitImpl = traitName
				}
				symbols = append(symbols, sym)
			}
		}
	}

	return symbols
}

func extractConst(node *sitter.Node, content []byte) languages.Symbol {
	nameNode := node.ChildByFieldName("name")
	name := ""
	if nameNode != nil {
		name = nameNode.Content(content)
	}

	vis := extractVisibility(node, content)
	doc := extractDoc(node, content)

	return &Const{
		name:       name,
		visibility: vis,
		doc:        doc,
		loc:        languages.NodeRange(node),
	}
}

func extractStatic(node *sitter.Node, content []byte) languages.Symbol {
	nameNode := node.ChildByFieldName("name")
	name := ""
	if nameNode != nil {
		name = nameNode.Content(content)
	}

	vis := extractVisibility(node, content)
	doc := extractDoc(node, content)

	return &Static{
		name:       name,
		visibility: vis,
		doc:        doc,
		loc:        languages.NodeRange(node),
	}
}

func extractTypeAlias(node *sitter.Node, content []byte) languages.Symbol {
	nameNode := node.ChildByFieldName("name")
	name := ""
	if nameNode != nil {
		name = nameNode.Content(content)
	}

	vis := extractVisibility(node, content)
	doc := extractDoc(node, content)

	return &TypeAlias{
		name:       name,
		visibility: vis,
		doc:        doc,
		loc:        languages.NodeRange(node),
	}
}

func extractMod(node *sitter.Node, content []byte) languages.Symbol {
	nameNode := node.ChildByFieldName("name")
	name := ""
	if nameNode != nil {
		name = nameNode.Content(content)
	}

	vis := extractVisibility(node, content)
	doc := extractDoc(node, content)

	return &Mod{
		name:       name,
		visibility: vis,
		doc:        doc,
		loc:        languages.NodeRange(node),
	}
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

func extractVisibility(node *sitter.Node, content []byte) string {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "visibility_modifier" {
			return child.Content(content)
		}
	}
	return ""
}

func extractDoc(node *sitter.Node, content []byte) string {
	prev := node.PrevNamedSibling()
	if prev == nil {
		return ""
	}

	// Rust doc comments can be line_comment (///) or block_comment (/** */)
	if prev.Type() != "line_comment" && prev.Type() != "block_comment" {
		return ""
	}

	commentEndLine := prev.EndPoint().Row
	declStartLine := node.StartPoint().Row
	if declStartLine-commentEndLine > 1 {
		return ""
	}

	text := prev.Content(content)

	// Handle /// doc comments
	if strings.HasPrefix(text, "///") {
		text = strings.TrimPrefix(text, "///")
	} else if strings.HasPrefix(text, "//!") {
		text = strings.TrimPrefix(text, "//!")
	} else if strings.HasPrefix(text, "/**") {
		text = strings.TrimPrefix(text, "/**")
		text = strings.TrimSuffix(text, "*/")
	}

	for line := range strings.SplitSeq(text, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}

	return ""
}
