package mag

import (
	"errors"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

type editListAction[T any] struct {
	Edit EditList[T]
}

// NewEditList creates a new ListAction that applies the given edit function to a list.
func NewEditList[T any](edit EditList[T]) ListAction {
	return &editListAction[T]{
		Edit: edit,
	}
}

// EditList returns changes to be applied to a list.
type EditList[T any] func(m *ListValue[T]) ([]Change, error)

type ListValue[T any] struct {
	List []*Node[T]
	Node *ast.SequenceNode
}

// Run edits keys and values of a given map.
func (a *editListAction[T]) Run(seq *ast.SequenceNode) error {
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
