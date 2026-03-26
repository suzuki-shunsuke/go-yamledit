package mag

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

type MapValue struct {
	Map       map[any]*KeyValue
	KeyValues []*KeyValue
	Node      *ast.MappingNode
}

type KeyValue struct {
	Key     any
	Value   any
	Comment string
	Node    *ast.MappingValueNode
	Index   int
}

// EditMapAction represents an action to edit a map key and value.
type EditMapAction struct {
	Edit EditMap
}

type EditMap func(m *MapValue, unmarshal func(any) error) ([]Change, error)

type Change interface {
	Run() error
}

// Run edits keys and values of a given map.
func (a *EditMapAction) Run(m *ast.MappingNode) error {
	mv := &MapValue{
		Map:       make(map[any]*KeyValue, len(m.Values)),
		KeyValues: make([]*KeyValue, 0, len(m.Values)),
		Node:      m,
	}
	mapIter := m.MapRange()
	idx := 0
	for mapIter.Next() {
		keyValue := mapIter.KeyValue()
		var k any
		if err := yaml.NodeToValue(keyValue.Key, &k); err != nil {
			return err
		}
		var v any
		if err := yaml.NodeToValue(keyValue.Key, &v); err != nil {
			return err
		}
		kv := &KeyValue{
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

type RenameKeyAction struct {
	Node *ast.MappingValueNode
	Key  any
}

func (a *RenameKeyAction) Run() error {
	comment := a.Node.Key.GetComment()
	v, err := yaml.ValueToNode(a.Key)
	if err != nil {
		return err
	}
	if err := v.SetComment(comment); err != nil {
		return fmt.Errorf("set comment to new key: %w", err)
	}
	k, ok := v.(ast.MapKeyNode)
	if !ok {
		return errors.New("failed to convert value to map key node")
	}
	a.Node.Key = k
	return nil
}
