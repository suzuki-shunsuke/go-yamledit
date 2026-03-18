package mag

import (
	"errors"
	"fmt"
	"slices"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// RemoveListItemAction represents an action to remove items from a sequence.
type RemoveListItemAction struct {
	// YAMLPath is a path to YAML sequence nodes that some items will be removed.
	// e.g. "$.reviewer"
	// https://github.com/goccy/go-yaml/blob/v1.19.2/path.go#L17-L22
	YAMLPath string
	// Remove chooses removed items.
	Remove RemoveListItem
}

// RemoveListItem returns indexes of items to be removed.
// If indexes is nil or empty, no item will be removed.
type RemoveListItem func(seq *ast.SequenceNode) ([]int, error)

// Run removes items from a node.
func (a *RemoveListItemAction) Run(node ast.Node) error {
	if a.YAMLPath == "" {
		return errors.New("yaml path is not set")
	}
	if a.Remove == nil {
		return errors.New("remove is not set")
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
	indexes, err := a.Remove(seq)
	if err != nil {
		return err
	}
	if len(indexes) == 0 {
		return nil
	}
	values := make([]ast.Node, 0, len(seq.Values)-len(indexes))
	for i, value := range seq.Values {
		if slices.Contains(indexes, i) {
			continue
		}
		values = append(values, value)
	}
	seq.Values = values
	return nil
}

type removeListItemsByIndexEditor struct {
	indexes []int
}

func (e *removeListItemsByIndexEditor) Remove(_ *ast.SequenceNode) ([]int, error) {
	return e.indexes, nil
}

// RemoveListItemsByIndex returns a RemoveListItem removing the item at the given indexes.
func RemoveListItemsByIndex(idxes ...int) RemoveListItem {
	s := &removeListItemsByIndexEditor{
		indexes: idxes,
	}
	return s.Remove
}
