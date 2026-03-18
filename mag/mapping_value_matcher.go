package mag

import (
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// MatchMappingValue returns true if a mapping value node matches.
type MatchMappingValue func(kv *ast.MappingValueNode) (bool, error)

// MatchMappingValueByKey returns a MatchMappingValue function that matches a mapping value node by keys.
func MatchMappingValueByKey(keys ...string) MatchMappingValue {
	return func(kv *ast.MappingValueNode) (bool, error) {
		var keyV any
		if err := yaml.NodeToValue(kv.Key, &keyV); err != nil {
			return false, err
		}
		for _, key := range keys {
			if compareKey(key, keyV) {
				return true, nil
			}
		}
		return false, nil
	}
}

// MatchAllMappingValues returns a MatchMappingValue function that matches all mapping value nodes.
func MatchAllMappingValues() MatchMappingValue {
	return func(_ *ast.MappingValueNode) (bool, error) {
		return true, nil
	}
}
