package mag

import (
	"slices"

	"github.com/goccy/go-yaml/ast"
)

// SelectItemFromList returns whether the item is selected.
type SelectItemFromList[T any] func(value *Node[T]) (bool, error)

// RemoveItemsFromList returns a ListAction removing items selected by the given function.
func RemoveItemsFromList[T any](remove SelectItemFromList[T]) ListAction {
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
			return RemoveItemsFromSequenceNode(m.Node, indexes...)
		},
	}
}

func RemoveItemsFromSequenceNode(node *ast.SequenceNode, indexes ...int) error {
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
