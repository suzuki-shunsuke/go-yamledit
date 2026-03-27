package mag

import (
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

type MapValue[K comparable, V any] struct {
	Map       map[K]*KeyValue[K]
	KeyValues []*KeyValue[K]
	Value     V
	Node      *ast.MappingNode
}

type KeyValue[K comparable] struct {
	Key     K
	Value   any
	Comment string
	Node    *ast.MappingValueNode
	Index   int
}

// EditMapAction represents an action to edit a map key and value.
type EditMapAction[K comparable, V any] struct {
	Edit EditMap[K, V]
}

type EditMap[K comparable, V any] func(m *MapValue[K, V]) ([]Change, error)

// Run edits keys and values of a given map.
func (a *EditMapAction[K, V]) Run(m *ast.MappingNode) error {
	mv := &MapValue[K, V]{
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
	edits, err := a.Edit(mv)
	if err != nil {
		return err
	}
	for _, edit := range edits {
		if err := edit.Run(); err != nil {
			return err
		}
	}
	return nil
}
