package mag

import (
	"slices"

	"github.com/goccy/go-yaml/ast"
)

// RemoveItemsFromMap returns a MapAction removing items from a map.
func RemoveItemsFromMap(match MatchMappingValue) MapAction {
	return &removeKeyAction{
		Match: match,
	}
}

// RemoveKeys returns a MapAction removing given keys from a map.
func RemoveKeys(keys ...any) MapAction {
	return &removeKeyAction{
		Match: MatchMappingValueByKey(keys...),
	}
}

type removeKeyAction struct {
	Match MatchMappingValue
}

func (a *removeKeyAction) Run(m *ast.MappingNode) error {
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
