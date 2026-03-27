package mag

import (
	"errors"
	"fmt"
	"slices"

	"github.com/goccy/go-yaml/ast"
)

// AddValuesToList returns an AddListItem adding the given value at the given index.
func AddValuesToList(idx int, values ...any) ListAction {
	return &EditListAction[any]{
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

// AddListItemFunc is a function that returns the value and index to insert into a list.
// If error is ErrNoop, no item will be added.
type AddListItemFunc func(seq *ast.SequenceNode) (any, int, error)

func AddListItemByFunc(fn AddListItemFunc) ListAction {
	return &addListItemAction{
		Add: fn,
	}
}

type addListItemAction struct {
	Add AddListItemFunc
}

func (a *addListItemAction) Run(seq *ast.SequenceNode) error {
	if a.Add == nil {
		return errors.New("Add is not set")
	}
	val, idx, err := a.Add(seq)
	if errors.Is(err, ErrNoop) {
		return nil
	}
	if err != nil {
		return err
	}
	v, err := valueToNode(val)
	if err != nil {
		return err
	}
	if idx < 0 {
		idx += len(seq.Values) + 1
	}
	seq.Values = append(seq.Values[:idx], append([]ast.Node{v}, seq.Values[idx:]...)...)
	return nil
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
