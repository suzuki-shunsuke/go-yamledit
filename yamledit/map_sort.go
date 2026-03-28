package yamledit

import (
	"slices"
)

type SortKeyFunc[K comparable] func(a, b *KeyValue[K]) int

// SortKey returns a MapAction sorting keys by the given function.
func SortKey[K comparable](fn SortKeyFunc[K]) MappingNodeAction {
	return &editMapAction[K, any]{
		Edit: func(m *Map[K, any]) error {
			kvs := make([]*KeyValue[K], len(m.KeyValues))
			copy(kvs, m.KeyValues)
			slices.SortStableFunc(kvs, fn)
			for i, item := range kvs {
				m.Node.Values[i] = item.Node
			}
			return nil
		},
	}
}
