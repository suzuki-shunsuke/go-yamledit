package mag

import (
	"errors"

	"github.com/goccy/go-yaml/ast"
)

// AddListItemFunc is a function that returns the value and index to insert into a list.
// If error is ErrNoop, no item will be added.
type AddListItemFunc func(seq *ast.SequenceNode) (any, int, error)

// AddValueToList returns an AddListItem adding the given value at the given index.
func AddValueToList(value any, idx int) ListAction {
	s := &staticAddListItemEditor{
		value: value,
		idx:   idx,
	}
	return &addListItemAction{Add: s.Add}
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

type staticAddListItemEditor struct {
	value any
	idx   int
}

func (e *staticAddListItemEditor) Add(_ *ast.SequenceNode) (any, int, error) {
	return e.value, e.idx, nil
}
