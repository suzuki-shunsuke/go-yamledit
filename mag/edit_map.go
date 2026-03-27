package mag

import (
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

type MapValue[T comparable] struct {
	Map       map[T]*KeyValue[T]
	KeyValues []*KeyValue[T]
	Node      *ast.MappingNode
}

type KeyValue[T comparable] struct {
	Key     T
	Value   any
	Comment string
	Node    *ast.MappingValueNode
	Index   int
}

// EditMapAction represents an action to edit a map key and value.
type EditMapAction[T comparable] struct {
	Edit EditMap[T]
}

type EditMap[T comparable] func(m *MapValue[T], unmarshal func(any) error) ([]Change, error)

type Change interface {
	Run() error
}

// Run edits keys and values of a given map.
func (a *EditMapAction[T]) Run(m *ast.MappingNode) error {
	mv := &MapValue[T]{
		Map:       make(map[T]*KeyValue[T], len(m.Values)),
		KeyValues: make([]*KeyValue[T], 0, len(m.Values)),
		Node:      m,
	}
	mapIter := m.MapRange()
	idx := 0
	for mapIter.Next() {
		keyValue := mapIter.KeyValue()
		var k T
		if err := yaml.NodeToValue(keyValue.Key, &k); err != nil {
			return err
		}
		var v T
		if err := yaml.NodeToValue(keyValue.Key, &v); err != nil {
			return err
		}
		kv := &KeyValue[T]{
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
	edits, err := a.Edit(mv, func(v any) error {
		return yaml.NodeToValue(m, &v)
	})
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
