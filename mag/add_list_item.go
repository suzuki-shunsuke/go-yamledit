package mag

import (
	"errors"
	"fmt"
	"slices"

	"github.com/goccy/go-yaml/ast"
)

// AddListItemFunc is a function that returns the value and index to insert into a list.
// If error is ErrNoop, no item will be added.
type AddListItemFunc func(seq *ast.SequenceNode) (any, int, error)

// AddValuesToList returns an AddListItem adding the given value at the given index.
func AddValuesToList(idx int, values ...any) ListAction {
	return &EditListAction[any]{
		Edit: func(m *ListValue[any]) ([]Change, error) {
			newIdx, err := checkInsertIndex(idx, len(m.List))
			if err != nil {
				return nil, err
			}
			nodes := make([]ast.Node, len(values))
			for i, v := range values {
				n, err := valueToNode(v)
				if err != nil {
					return nil, fmt.Errorf("convert value to node: %w", err)
				}
				nodes[i] = n
			}
			return []Change{
				&ChangeAddItemsToList{
					List:  m.Node,
					Index: newIdx,
					Nodes: nodes,
				},
			}, nil
		},
	}
}

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
	List  *ast.SequenceNode
	Nodes []ast.Node
	Index int
}

func (a *ChangeAddItemsToList) Run() error {
	a.List.Values = slices.Insert(a.List.Values, a.Index, a.Nodes...)
	return nil
}
