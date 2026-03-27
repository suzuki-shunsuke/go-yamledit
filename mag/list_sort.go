package mag

import (
	"errors"
	"slices"

	"github.com/goccy/go-yaml/ast"
)

// Node represents a YAML node.
type Node[T any] struct {
	Node ast.Node
	// Value is the value of the node.
	Value T
	// Comment is the comment of the node.
	Comment string
}

type SortFunc[T any] func(a, b *Node[T]) int

func SortList[T any](fn SortFunc[T]) SequenceNodeAction {
	return &editListAction[T]{
		Edit: func(m *List[T]) error {
			if fn == nil {
				return errors.New("sort function is nil")
			}

			values := make([]*Node[T], len(m.List))
			copy(values, m.List)
			slices.SortStableFunc(values, fn)

			nodes := make([]ast.Node, len(m.List))
			for i, value := range values {
				nodes[i] = value.Node
			}

			m.Node.Values = nodes
			return nil
		},
	}
}
