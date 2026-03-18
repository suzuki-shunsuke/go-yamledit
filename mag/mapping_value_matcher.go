package mag

import (
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

type MatchMappingValue func(kv *ast.MappingValueNode) (bool, error)

type keyMVMatcher struct {
	key string
}

func MatchMappingValueByKey(key string) MatchMappingValue {
	m := &keyMVMatcher{key: key}
	return m.Match
}

func (m *keyMVMatcher) Match(kv *ast.MappingValueNode) (bool, error) {
	var keyV any
	if err := yaml.NodeToValue(kv.Key, &keyV); err != nil {
		return false, err
	}
	return compareKey(m.key, keyV), nil
}

func MatchAllMappingValues() MatchMappingValue {
	m := &allMVMatcher{}
	return m.Match
}

type allMVMatcher struct{}

func (m *allMVMatcher) Match(_ *ast.MappingValueNode) (bool, error) {
	return true, nil
}
