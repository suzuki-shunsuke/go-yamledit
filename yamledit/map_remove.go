package yamledit

import (
	"slices"

	"github.com/goccy/go-yaml/ast"
)

// RemoveKeys returns a MapAction removing given keys from a map.
func RemoveKeys(keys ...any) MappingNodeAction {
	return &editMapAction[any, any]{
		Edit: func(m *Map[any, any]) error {
			indexes := make([]int, 0, len(keys))
			for _, key := range keys {
				kv, ok := m.Map[key]
				if !ok {
					continue
				}
				indexes = append(indexes, kv.Index)
			}
			return RemoveKeysFromMappingNode(m.Node, indexes...)
		},
	}
}

func RemoveKeysFromMappingNode(node *ast.MappingNode, indexes ...int) error {
	values := make([]*ast.MappingValueNode, 0, len(node.Values)-len(indexes))
	for i, v := range node.Values {
		if slices.Contains(indexes, i) {
			continue
		}
		values = append(values, v)
	}
	node.Values = values
	return nil
}
