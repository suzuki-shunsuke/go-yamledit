package mag

import (
	"errors"
	"fmt"
	"slices"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// SortListAction represents an action to sort lists.
type SortListAction[T any] struct {
	// YAMLPath is a path to YAML sequence nodes that new items will be sorted.
	// e.g. "$.reviewers"
	// https://github.com/goccy/go-yaml/blob/v1.19.2/path.go#L17-L22
	YAMLPath string
	// Sort is a function to sort list.
	Sort func(a, b *Item[T]) int
}

type Item[T any] struct {
	Node    ast.Node
	Value   T
	Comment string
}

// SortList is a function to sort list.
// This is compatible with slices.SortStableFunc.
// https://pkg.go.dev/slices#SortStableFunc
type SortList[T any] func(a, b *Item[T]) int

// Run sorts lists.
func (a *SortListAction[T]) Run(node ast.Node) error {
	if a.Sort == nil {
		return errors.New("sort is not set")
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
		if err := a.sort(elem); err != nil {
			return err
		}
	}
	return nil
}

func (a *SortListAction[T]) sort(elem ast.Node) error {
	seq, ok := elem.(*ast.SequenceNode)
	if !ok {
		return fmt.Errorf("expected a sequence node: %s", elem.Type().String())
	}

	var values []T
	if err := yaml.NodeToValue(seq, &values); err != nil {
		return err
	}
	valueWithNodes := make([]*Item[T], len(values))
	for i, value := range values {
		valueWithNodes[i] = &Item[T]{
			Node:  seq.Values[i],
			Value: value,
		}
	}
	slices.SortStableFunc(valueWithNodes, a.Sort)

	for i, item := range valueWithNodes {
		seq.Values[i] = item.Node
	}
	return nil
}
