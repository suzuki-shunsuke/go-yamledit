package mag

import (
	"errors"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// EditListAction represents an action to edit a map key and value.
type EditListAction[T any] struct {
	Edit EditList[T]
}

type EditList[T any] func(m *ListValue[T]) ([]Change, error)

type ListValue[T any] struct {
	List []*Node[T]
	Node *ast.SequenceNode
}

// Run edits keys and values of a given map.
func (a *EditListAction[T]) Run(seq *ast.SequenceNode) error {
	if a.Edit == nil {
		return errors.New("edit function is nil")
	}
	mv := &ListValue[T]{
		List: make([]*Node[T], len(seq.Values)),
		Node: seq,
	}
	for i, value := range seq.Values {
		var v T
		if err := yaml.NodeToValue(value, &v); err != nil {
			return err
		}
		mv.List[i] = &Node[T]{
			Value:   v,
			Node:    value,
			Comment: getComment(value),
		}
	}

	edits, err := a.Edit(mv)
	if err != nil {
		return err
	}
	for _, edit := range edits {
		if err := edit.Run(); err != nil {
			return err
		}
	}
	return nil
}
