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
	Sort func(a, b *Node[T]) int
}

// Node represents a YAML node.
type Node[T any] struct {
	Node ast.Node
	// Value is the value of the node.
	Value T
	// Comment is the comment of the node.
	Comment string
}

// Run sorts the given sequence.
func (a *SortListAction[T]) Run(seq *ast.SequenceNode) error {
	if a.Sort == nil {
		return errors.New("sort is not set")
	}
	var values []T
	if err := yaml.NodeToValue(seq, &values); err != nil {
		return err
	}
	valueWithNodes := make([]*Node[T], len(values))
	for i, value := range values {
		valueWithNodes[i] = &Node[T]{
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
