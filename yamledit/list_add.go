package yamledit

import (
	"fmt"
	"slices"

	"github.com/goccy/go-yaml/ast"
)

// AddValuesToList returns an AddListItem adding the given value at the given index.
func AddValuesToList(idx int, values ...any) SequenceNodeAction {
	return EditListAction[any](
		func(m *List[any]) error {
			return AddValuesToSequenceNode(m.Node, idx, values...)
		},
	)
}

func AddValuesToSequenceNode(seq *ast.SequenceNode, index int, values ...any) error {
	idx, err := checkInsertIndex(index, len(seq.Values))
	if err != nil {
		return err
	}
	nodes := make([]ast.Node, 0, len(values))
	for _, v := range values {
		if b, ok := v.(*Bytes); ok {
			if b.isList {
				n, err := BytesToNode(b.b)
				if err != nil {
					return fmt.Errorf("convert bytes to node: %w", err)
				}
				seq, ok := n.(*ast.SequenceNode)
				if !ok {
					return fmt.Errorf("expected sequence node, got %T", n)
				}
				nodes = append(nodes, seq.Values...)
				continue
			}
			n, err := BytesToNode(b.b)
			if err != nil {
				return fmt.Errorf("convert bytes to node: %w", err)
			}
			nodes = append(nodes, n)
			continue
		}
		n, err := valueToNode(v)
		if err != nil {
			return fmt.Errorf("convert value to node: %w", err)
		}
		nodes = append(nodes, n)
	}
	seq.Values = slices.Insert(seq.Values, idx, nodes...)
	return nil
}
