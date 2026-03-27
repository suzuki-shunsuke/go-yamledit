package mag

import (
	"slices"

	"github.com/goccy/go-yaml/ast"
)

// SelectValue returns whether the value is selected.
type SelectValue[T any] func(value *Node[T]) (bool, error)

// RemoveValuesFromList returns a ListAction removing items selected by the given function.
func RemoveValuesFromList[T any](remove SelectValue[T]) ListAction {
	return &editListAction[T]{
		Edit: func(m *ListValue[T]) error {
			indexes := make([]int, 0, len(m.List))
			for i, node := range m.List {
				f, err := remove(node)
				if err != nil {
					return err
				}
				if f {
					indexes = append(indexes, i)
				}
			}
			return RemoveValuesFromSequenceNode(m.Node, indexes...)
		},
	}
}

func RemoveValuesFromSequenceNode(node *ast.SequenceNode, indexes ...int) error {
	values := make([]ast.Node, 0, len(node.Values)-len(indexes))
	if err := normalizeIndexes(indexes, len(node.Values)); err != nil {
		return err
	}
	for i, value := range node.Values {
		if slices.Contains(indexes, i) {
			continue
		}
		values = append(values, value)
	}
	node.Values = values
	return nil
}
