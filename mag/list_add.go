package mag

import (
	"fmt"
	"slices"

	"github.com/goccy/go-yaml/ast"
)

// AddValuesToList returns an AddListItem adding the given value at the given index.
func AddValuesToList(idx int, values ...any) ListAction {
	return &editListAction[any]{
		Edit: func(m *ListValue[any]) error {
			return AddValuesToSequenceNode(m.Node, idx, values...)
		},
	}
}

func AddValuesToSequenceNode(seq *ast.SequenceNode, index int, values ...any) error {
	idx, err := checkInsertIndex(index, len(seq.Values))
	if err != nil {
		return err
	}
	nodes := make([]ast.Node, len(values))
	for i, v := range values {
		n, err := valueToNode(v)
		if err != nil {
			return fmt.Errorf("convert value to node: %w", err)
		}
		nodes[i] = n
	}
	seq.Values = slices.Insert(seq.Values, idx, nodes...)
	return nil
}
