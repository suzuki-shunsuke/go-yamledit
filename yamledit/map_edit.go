package yamledit

import (
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// Map represents a map.
// K is the type of keys.
// V is the type of the map.
type Map[K comparable, V any] struct {
	Map       map[K]*KeyValue[K]
	KeyValues []*KeyValue[K]
	Value     V
	Node      *ast.MappingNode
}

// KeyValue represents a key-value pair in a map.
type KeyValue[K comparable] struct {
	Key     K
	Value   any
	Comment string
	Node    *ast.MappingValueNode
	Index   int
}

func EditMapAction[K comparable, V any](edit EditMap[K, V]) MappingNodeAction {
	return &editMapAction[K, V]{Edit: edit}
}

type EditMap[K comparable, V any] func(m *Map[K, V]) error

// editMapAction represents an action to edit a map key and value.
type editMapAction[K comparable, V any] struct {
	Edit EditMap[K, V]
}

// Run edits keys and values of a given map.
func (a *editMapAction[K, V]) Run(m *ast.MappingNode) error {
	mv := &Map[K, V]{
		Map:       make(map[K]*KeyValue[K], len(m.Values)),
		KeyValues: make([]*KeyValue[K], 0, len(m.Values)),
		Node:      m,
	}
	mapIter := m.MapRange()
	idx := 0

	var value V
	if err := yaml.NodeToValue(m, &value); err != nil {
		return err
	}
	mv.Value = value

	for mapIter.Next() {
		keyValue := mapIter.KeyValue()
		var k K
		if err := yaml.NodeToValue(keyValue.Key, &k); err != nil {
			return err
		}
		var v any
		if err := yaml.NodeToValue(keyValue.Value, &v); err != nil {
			return err
		}
		kv := &KeyValue[K]{
			Key:     k,
			Value:   v,
			Node:    keyValue,
			Comment: getComment(keyValue.Value),
			Index:   idx,
		}
		mv.Map[k] = kv
		mv.KeyValues = append(mv.KeyValues, kv)
		idx++
	}

	if err := a.Edit(mv); err != nil {
		return err
	}
	return nil
}
