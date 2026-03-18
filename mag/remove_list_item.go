package mag

import (
	"errors"
	"slices"

	"github.com/goccy/go-yaml/ast"
)

// SelectItemsFromList returns indexes of items to be removed.
// If indexes is nil or empty, no item will be removed.
type SelectItemsFromList func(seq *ast.SequenceNode) ([]int, error)

// RemoveListItemsByIndex returns a ListAction removing items at the given indexes.
func RemoveListItemsByIndex(indexes ...int) ListAction {
	return &removeListItemAction{
		Remove: func(_ *ast.SequenceNode) ([]int, error) {
			return indexes, nil
		},
	}
}

// RemoveItemsFromList returns a ListAction removing items selected by the given function.
func RemoveItemsFromList(remove SelectItemsFromList) ListAction {
	return &removeListItemAction{
		Remove: remove,
	}
}

type removeListItemAction struct {
	Remove SelectItemsFromList
}

func (a *removeListItemAction) Run(seq *ast.SequenceNode) error {
	if a.Remove == nil {
		return errors.New("remove is not set")
	}
	indexes, err := a.Remove(seq)
	if err != nil {
		return err
	}
	if len(indexes) == 0 {
		return nil
	}
	values := make([]ast.Node, 0, len(seq.Values)-len(indexes))
	for i, value := range seq.Values {
		if slices.Contains(indexes, i) {
			continue
		}
		values = append(values, value)
	}
	seq.Values = values
	return nil
}
