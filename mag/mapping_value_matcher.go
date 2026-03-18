package mag

import (
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// MatchMappingValue returns true if a mapping value node matches.
type MatchMappingValue func(kv *ast.MappingValueNode) (bool, error)

type keyMVMatcher struct {
	keys []string
}

// MatchMappingValueByKey returns a MatchMappingValue function that matches a mapping value node by keys.
func MatchMappingValueByKey(keys ...string) MatchMappingValue {
	m := &keyMVMatcher{keys: keys}
	return m.Match
}

func (m *keyMVMatcher) Match(kv *ast.MappingValueNode) (bool, error) {
	var keyV any
	if err := yaml.NodeToValue(kv.Key, &keyV); err != nil {
		return false, err
	}
	for _, key := range m.keys {
		if compareKey(key, keyV) {
			return true, nil
		}
	}
	return false, nil
}

// MatchAllMappingValues returns a MatchMappingValue function that matches all mapping value nodes.
func MatchAllMappingValues() MatchMappingValue {
	m := &allMVMatcher{}
	return m.Match
}

type allMVMatcher struct{}

func (m *allMVMatcher) Match(_ *ast.MappingValueNode) (bool, error) {
	return true, nil
}
