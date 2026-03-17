package mag

import "github.com/goccy/go-yaml/ast"

type MappingValueEditor interface {
	Edit(mv *ast.MappingValueNode) (any, any, error)
}

type FixedEditor struct {
	key   any
	value any
}

func NewFixedEditor(key, value any) *FixedEditor {
	return &FixedEditor{
		key:   key,
		value: value,
	}
}

func (f *FixedEditor) Edit(_ *ast.MappingValueNode) (any, any, error) {
	return f.key, f.value, nil
}
