package languages

import (
	sitter "github.com/smacker/go-tree-sitter"
)

// NodeRange converts a tree-sitter node to a Range
func NodeRange(node *sitter.Node) Range {
	start := node.StartPoint()
	end := node.EndPoint()
	return Range{
		Start: Position{Line: int(start.Row), Character: int(start.Column)},
		End:   Position{Line: int(end.Row), Character: int(end.Column)},
	}
}
