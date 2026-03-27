package mag

import (
	"slices"

	"github.com/goccy/go-yaml/ast"
)

type SortKeyFunc[K comparable] func(a, b *KeyValue[K]) int

// SortKey returns a MapAction sorting keys by the given function.
func SortKey[K comparable](fn SortKeyFunc[K]) MapAction {
	return &EditMapAction[K, any]{
		Edit: func(m *MapValue[K, any]) ([]Change, error) {
			kvs := make([]*KeyValue[K], len(m.KeyValues))
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
