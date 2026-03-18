package mag

import (
	"errors"
	"slices"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// SortListAction represents an action to sort lists.
type SortListAction[T any] struct {
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

func (a *SortListAction[T]) Run(seq *ast.SequenceNode) error {
	if a.Sort == nil {
		return errors.New("sort is not set")
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
