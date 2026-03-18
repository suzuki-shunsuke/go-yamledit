package mag

import "github.com/goccy/go-yaml/ast"

// Action represents an operation updating YAML AST nodes.
type Action interface {
	// Run modifies the given YAML AST node.
	Run(node ast.Node) error
}
