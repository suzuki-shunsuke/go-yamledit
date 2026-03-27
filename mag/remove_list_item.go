package mag

import (
	"slices"

	"github.com/goccy/go-yaml/ast"
)

// SelectItemsFromList returns indexes of items to be removed.
// If indexes is nil or empty, no item will be removed.
type SelectItemFromList[T any] func(value *Node[T]) (bool, error)

// RemoveItemsFromList returns a ListAction removing items selected by the given function.
func RemoveItemsFromList[T any](remove SelectItemFromList[T]) ListAction {
	return &EditListAction[T]{
		Edit: func(m *ListValue[T]) ([]Change, error) {
			indexes := make([]int, 0, len(m.List))
			for i, node := range m.List {
				f, err := remove(node)
				if err != nil {
					return nil, err
				}
				if f {
					indexes = append(indexes, i)
				}
			}
			return []Change{
				&ChangeRemoveItemFromList{
					Node:    m.Node,
					Indexes: indexes,
				},
			}, nil
		},
	}
}

type ChangeRemoveItemFromList struct {
	Node    *ast.SequenceNode
	Indexes []int
}

func (a *ChangeRemoveItemFromList) Run() error {
	values := make([]ast.Node, 0, len(a.Node.Values)-len(a.Indexes))
	if err := normalizeIndexes(a.Indexes, len(a.Node.Values)); err != nil {
		return err
	}
	for i, value := range a.Node.Values {
		if slices.Contains(a.Indexes, i) {
			continue
		}
		values = append(values, value)
	}
	a.Node.Values = values
	return nil
}
