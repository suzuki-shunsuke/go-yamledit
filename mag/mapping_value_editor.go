package mag

import (
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

type MappingValueEditor interface {
	Edit(mv *ast.MappingValueNode) (any, any, error)
}

func NewStaticMappingValueEditor(key, value any) MappingValueEditor {
	return &generalMappingValueEditor{
		edit: func(_ *ast.MappingValueNode, _ *MappingValue) (any, any, error) {
			return key, value, nil
		},
	}
}

type MappingValue struct {
	Key          any
	Value        any
	KeyComment   string
	ValueComment string
}

type generalMappingValueEditor struct {
	edit func(node *ast.MappingValueNode, mv *MappingValue) (any, any, error)
}

func NewGeneralMappingValueEditor(edit func(node *ast.MappingValueNode, mv *MappingValue) (any, any, error)) MappingValueEditor {
	return &generalMappingValueEditor{
		edit: edit,
	}
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
		Key:          kv,
		Value:        value,
		KeyComment:   getComment(node.Key),
		ValueComment: getComment(node.Value),
	})
}

func getComment(node ast.Node) string {
	if node == nil {
		return ""
	}
	cn := node.GetComment()
	if cn == nil {
		return ""
	}
	return cn.String()
}
