package mag

import (
	"fmt"
	"slices"

	"github.com/goccy/go-yaml/ast"
)

// AddValuesToList returns an AddListItem adding the given value at the given index.
func AddValuesToList(idx int, values ...any) ListAction {
	return &editListAction[any]{
		Edit: func(m *ListValue[any]) ([]Change, error) {
			return []Change{
				&ChangeAddItemsToList{
					List:   m.Node,
					Index:  idx,
					Values: values,
				},
			}, nil
		},
	}
}

type ChangeAddItemsToList struct {
	List   *ast.SequenceNode
	Values []any
	Index  int
}

func (a *ChangeAddItemsToList) Run() error {
	idx, err := checkInsertIndex(a.Index, len(a.List.Values))
	if err != nil {
		return err
	}
	nodes := make([]ast.Node, len(a.Values))
	for i, v := range a.Values {
		n, err := valueToNode(v)
		if err != nil {
			return fmt.Errorf("convert value to node: %w", err)
		}
		nodes[i] = n
	}
	a.List.Values = slices.Insert(a.List.Values, idx, nodes...)
	return nil
}
