package mag

import (
	"errors"
	"fmt"
	"slices"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

type RemoveListItemAction struct {
	YAMLPath string
	Remove   RemoveListItem
	Depth    int
}

type RemoveListItem func(seq *ast.SequenceNode) (int, error)

func (a *RemoveListItemAction) Run(node ast.Node) error {
	if a.Remove == nil {
		return errors.New("remove is not set")
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
		if err := a.remove(elem); err != nil {
			return err
		}
	}
	return nil
}

func (a *RemoveListItemAction) remove(elem ast.Node) error {
	seq, ok := elem.(*ast.SequenceNode)
	if !ok {
		return fmt.Errorf("expected a sequence node: %s", elem.Type().String())
	}
	idx, err := a.Remove(seq)
	if errors.Is(err, ErrNoop) {
		return nil
	}
	if err != nil {
		return err
	}
	seq.Values = slices.Delete(seq.Values, idx, idx+1)
	return nil
}

type staticRemoveListItemEditor struct {
	idx int
}

func (e *staticRemoveListItemEditor) Remove(_ *ast.SequenceNode) (int, error) {
	return e.idx, nil
}

func NewStaticRemoveListItemEditor(idx int) RemoveListItem {
	s := &staticRemoveListItemEditor{
		idx: idx,
	}
	return s.Remove
}
