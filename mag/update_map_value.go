package mag

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

type UpdateMapValueAction struct {
	YAMLPath string
	Matcher  MappingValueMatcher
	Value    any
}

func (a *UpdateMapValueAction) Run(node ast.Node) error {
	if a.Matcher == nil {
		return errors.New("matcher is not set")
	}
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

		f, err := a.Matcher.Match(keyValue)
		if err != nil {
			return err
		}
		if !f {
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
