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

// AddListItemAction represents an action to add a list item to a YAML sequence node.
type AddListItemAction struct {
	// YAMLPath is a path to YAML sequence nodes that new items will be added to.
	// e.g. "$.reviewers"
	// https://github.com/goccy/go-yaml/blob/v1.19.2/path.go#L17-L22
	YAMLPath string
	// Add is a function that returns the value and index to insert into the sequence.
	Add AddListItem
}

// AddListItem is a function that returns the value and index to insert into a list.
// If error is ErrNoop, no item will be added.
type AddListItem func(seq *ast.SequenceNode) (any, int, error)

// Run adds a list item to a YAML sequence node.
func (a *AddListItemAction) Run(node ast.Node) error {
	if a.Add == nil {
		return errors.New("Add is not set")
	}
	if a.YAMLPath == "" {
		return errors.New("YAMLPath is not set")
	}
	path, err := yaml.PathString(a.YAMLPath)
	if err != nil {
		return fmt.Errorf("parse a YAML path: %w", err)
	}
	n, err := path.FilterNode(node)
	if err != nil {
		return fmt.Errorf("filter node by YAML Path: %w", err)
	}
	nodes, err := flatten(n, getDepthByPath(a.YAMLPath))
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

// NewStaticAddListItemEditor returns an AddListItem adding the given value at the given index.
func NewStaticAddListItemEditor(value any, idx int) AddListItem {
	s := &staticAddListItemEditor{
		value: value,
		idx:   idx,
	}
	return s.Add
}
