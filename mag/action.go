package mag

import "github.com/goccy/go-yaml/ast"

type Action interface {
	Run(node ast.Node) error
}
