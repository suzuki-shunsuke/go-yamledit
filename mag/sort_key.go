package mag

import (
	"slices"

	"github.com/goccy/go-yaml/ast"
)

type SortKeyFunc[T comparable] func(a, b *KeyValue[T]) int

// SortKey returns a MapAction sorting keys by the given function.
func SortKey[T comparable](fn SortKeyFunc[T]) MapAction {
	return &EditMapAction[T]{
		Edit: func(m *MapValue[T], _ func(any) error) ([]Change, error) {
			kvs := make([]*KeyValue[T], len(m.KeyValues))
			copy(kvs, m.KeyValues)
			slices.SortStableFunc(kvs, fn)
			values := make([]*ast.MappingValueNode, len(m.KeyValues))
			for i, item := range kvs {
				values[i] = item.Node
			}
			return []Change{
				&ChangeSortKey{
					Node:   m.Node,
					Values: values,
				},
			}, nil
		},
	}
}

type ChangeSortKey struct {
	Node   *ast.MappingNode
	Values []*ast.MappingValueNode
}

func (a *ChangeSortKey) Run() error {
	a.Node.Values = a.Values
	return nil
}
