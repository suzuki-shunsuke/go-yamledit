package mag

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// TODO
// slice
//   [x] Add element to a slice
//   [x] Remove element from a slice
//   [ ] Sort slice

type AddListItemAction struct {
	YAMLPath string
	Add      AddListItem
	Depth    int
}

type AddListItem func(seq *ast.SequenceNode) (any, int, error)

func (a *AddListItemAction) Run(node ast.Node) error {
	if a.Add == nil {
		return errors.New("add is not set")
	}
	if a.Depth < 0 {
		return fmt.Errorf("depth must be >= 0: %d", a.Depth)
	}
	path, err := yaml.PathString(a.YAMLPath)
	if err != nil {
		return fmt.Errorf("parse a YAML path: %w", err)
	}
	n, err := path.FilterNode(node)
	if err != nil {
		return fmt.Errorf("filter node by YAML Path: %w", err)
	}
	nodes, err := flatten(n, a.Depth)
	if err != nil {
		return err
	}
	for _, elem := range nodes {
		if err := a.add(elem); err != nil {
			return err
		}
	}
	return nil
}

func (a *AddListItemAction) add(elem ast.Node) error {
	seq, ok := elem.(*ast.SequenceNode)
	if !ok {
		return fmt.Errorf("expected a sequence node: %s", elem.Type().String())
	}
	val, idx, err := a.Add(seq)
	if err != nil {
		return err
	}
	if !IsChanged(val) {
		return nil
	}
	v, err := yaml.ValueToNode(val)
	if err != nil {
		return err
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

func NewStaticAddListItemEditor(value any, idx int) AddListItem {
	s := &staticAddListItemEditor{
		value: value,
		idx:   idx,
	}
	return s.Add
}
