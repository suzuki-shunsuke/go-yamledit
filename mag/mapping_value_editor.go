package mag

import (
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// EditMappingValue returns a key and value to be set for the mapping value node.
// The first return value is the key, the second is the value.
// If NoChange is returned, key or value is not changed.
// NoChange is used to change only one of the key or value.
type EditMappingValue func(mv *ast.MappingValueNode) (any, any, error)

// EditMappingValueStatic returns a MappingValueEditor editing a mapping key and value to the given key and value.
// Matcher must choose only one pair of key and value.
func EditMappingValueStatic(key, value any) EditMappingValue {
	e := &generalMappingValueEditor{
		edit: func(_ *ast.MappingValueNode, _ *MappingValue) (any, any, error) {
			return key, value, nil
		},
	}
	return e.Edit
}

// MappingValue represents a mapping key and value.
type MappingValue struct {
	Key     any
	Value   any
	Comment string
}

type generalMappingValueEditor struct {
	edit func(node *ast.MappingValueNode, mv *MappingValue) (any, any, error)
}

// NewEditMappingValue returns a MappingValueEditor editing a mapping key and value using the given edit function.
func NewEditMappingValue(edit func(node *ast.MappingValueNode, mv *MappingValue) (any, any, error)) EditMappingValue {
	e := &generalMappingValueEditor{
		edit: edit,
	}
	return e.Edit
}

func (f *generalMappingValueEditor) Edit(node *ast.MappingValueNode) (any, any, error) {
	var kv any
	if err := yaml.NodeToValue(node.Key, &kv); err != nil {
		return nil, nil, err
	}
	var value any
	if err := yaml.NodeToValue(node.Key, &value); err != nil {
		return nil, nil, err
	}
	return f.edit(node, &MappingValue{
		Key:     kv,
		Value:   value,
		Comment: getComment(node.Value),
	})
}
