package mag

import (
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

type MappingValueMatcher interface {
	Match(kv *ast.MappingValueNode) (bool, error)
}

type KeyMVMatcher struct {
	key string
}

func NewKeyMVMatcher(key string) *KeyMVMatcher {
	return &KeyMVMatcher{key: key}
}

func (m *KeyMVMatcher) Match(kv *ast.MappingValueNode) (bool, error) {
	var keyV any
	if err := yaml.NodeToValue(kv.Key, &keyV); err != nil {
		return false, err
	}
	return compareKey(m.key, keyV), nil
}

type AllMVMatcher struct{}

func (m *AllMVMatcher) Match(_ *ast.MappingValueNode) (bool, error) {
	return true, nil
}
