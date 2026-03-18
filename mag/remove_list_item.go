package mag

import (
	"errors"
	"slices"

	"github.com/goccy/go-yaml/ast"
)

// RemoveListItemAction represents an action to remove items from a sequence.
type RemoveListItemAction struct {
	// Remove chooses removed items.
	Remove RemoveListItem
}

// RemoveListItem returns indexes of items to be removed.
// If indexes is nil or empty, no item will be removed.
type RemoveListItem func(seq *ast.SequenceNode) ([]int, error)

// Run removes items from the given sequence.
func (a *RemoveListItemAction) Run(seq *ast.SequenceNode) error {
	if a.Remove == nil {
		return errors.New("Remove is not set")
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

type removeListItemsByIndexEditor struct {
	indexes []int
}

func (e *removeListItemsByIndexEditor) Remove(_ *ast.SequenceNode) ([]int, error) {
	return e.indexes, nil
}

// RemoveListItemsByIndex returns a ListAction removing items at the given indexes.
func RemoveListItemsByIndex(idxes ...int) ListAction {
	s := &removeListItemsByIndexEditor{
		indexes: idxes,
	}
	return &RemoveListItemAction{
		Remove: s.Remove,
	}
}
