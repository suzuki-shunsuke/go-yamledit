package mag

import (
	"fmt"
	"slices"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

type RemoveKeyAction struct {
	YAMLPath string
	Matcher  MappingValueMatcher
}

func (a *RemoveKeyAction) Run(node ast.Node) error {
	path, err := yaml.PathString(a.YAMLPath)
	if err != nil {
		return fmt.Errorf("parse a YAML path: %w", err)
	}
	n, err := path.FilterNode(node)
	if err != nil {
		return fmt.Errorf("filter node by YAML Path: %w", err)
	}
	switch v := n.(type) {
	case *ast.MappingNode:
		return a.removeKeyFromMap(v)
	case *ast.SequenceNode:
		for _, elem := range v.Values {
			m, ok := elem.(*ast.MappingNode)
			if !ok {
				continue
			}
			if err := a.removeKeyFromMap(m); err != nil {
				return err
			}
		}
		return nil
	default:
		return nil
	}
}

func (a *RemoveKeyAction) removeKeyFromMap(m *ast.MappingNode) error {
	idx := 0
	mapIter := m.MapRange()
	for mapIter.Next() {
		f, err := a.Matcher.Match(mapIter.KeyValue())
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
