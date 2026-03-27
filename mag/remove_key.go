package mag

import (
	"slices"

	"github.com/goccy/go-yaml/ast"
)

// RemoveKeys returns a MapAction removing given keys from a map.
func RemoveKeys(keys ...any) MapAction {
	return &EditMapAction{
		Edit: func(m *MapValue, _ func(any) error) ([]Change, error) {
			indexes := make([]int, 0, len(keys))
			for _, key := range keys {
				kv, ok := m.Map[key]
				if !ok {
					continue
				}
				indexes = append(indexes, kv.Index)
			}
			return []Change{
				&ChangeRemoveKeys{
					Indexes: indexes,
					Node:    m.Node,
				},
			}, nil
		},
	}
}

type ChangeRemoveKeys struct {
	Node    *ast.MappingNode
	Indexes []int
}

func (a *ChangeRemoveKeys) Run() error {
	values := make([]*ast.MappingValueNode, 0, len(a.Node.Values)-len(a.Indexes))
	for i, v := range a.Node.Values {
		if slices.Contains(a.Indexes, i) {
			continue
		}
		values = append(values, v)
	}
	a.Node.Values = values
	return nil
}
