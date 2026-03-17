package mag

import (
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

type UpdateMapValueAction struct {
	YAMLPath string
	Key      any
	Value    any
}

func (a *UpdateMapValueAction) Run(node ast.Node) error {
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
		return a.updateMapValue(v)
	case *ast.SequenceNode:
		for _, elem := range v.Values {
			m, ok := elem.(*ast.MappingNode)
			if !ok {
				continue
			}
			if err := a.updateMapValue(m); err != nil {
				return err
			}
		}
		return nil
	default:
		return nil
	}
}

func (a *UpdateMapValueAction) updateMapValue(m *ast.MappingNode) error {
	mapIter := m.MapRange()
	for mapIter.Next() {
		keyValue := mapIter.KeyValue()
		var keyV any
		if err := yaml.NodeToValue(keyValue.Key, &keyV); err != nil {
			return err
		}
		if !compareKey(a.Key, keyV) {
			continue
		}
		n, err := yaml.ValueToNode(a.Value)
		if err != nil {
			return err
		}
		comment := keyValue.Value.GetComment()
		keyValue.Value = n
		if err := keyValue.Value.SetComment(comment); err != nil {
			return err
		}
		return nil
	}
	return nil
}
