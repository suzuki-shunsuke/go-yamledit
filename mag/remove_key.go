package mag

import (
	"slices"

	"github.com/goccy/go-yaml/ast"
)

// RemoveKeyAction represents an action to remove keys from a map.
type RemoveKeyAction struct {
	// Match filters mapping keys and values to be removed.
	Match MatchMappingValue
}

// Run removes keys from the given map.
func (a *RemoveKeyAction) Run(m *ast.MappingNode) error {
	idx := 0
	mapIter := m.MapRange()
	for mapIter.Next() {
		f, err := a.Match(mapIter.KeyValue())
		if err != nil {
			return err
		}
		if !f {
			idx++
			continue
		}
		m.Values = slices.Delete(m.Values, idx, idx+1)
		return nil
	}
	return nil
}

// RemoveKeys returns a MapAction removing given keys from a map.
func RemoveKeys(keys ...string) MapAction {
	return &RemoveKeyAction{
		Match: MatchMappingValueByKey(keys...),
	}
}
