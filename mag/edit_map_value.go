package mag

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

type EditMapValueAction struct {
	YAMLPath string
	Matcher  MappingValueMatcher
	Editor   MappingValueEditor
}

func (a *EditMapValueAction) Run(node ast.Node) error {
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

	nodes, err := flatten(n, -1)
	if err != nil {
		return err
	}
	for _, elem := range nodes {
		e, ok := elem.(*ast.MappingNode)
		if !ok {
			return fmt.Errorf("element is not a mapping node: %s", elem.Type().String())
		}
		if err := a.editMapValue(e); err != nil {
			return err
		}
	}
	return nil
}

func (a *EditMapValueAction) editMapValue(m *ast.MappingNode) error {
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

		newKey, newValue, err := a.Editor.Edit(keyValue)
		if err != nil {
			return err
		}
		if IsChanged(newKey) {
			if err := a.editKey(keyValue, newKey); err != nil {
				return fmt.Errorf("edit key: %w", err)
			}
		}
		if IsChanged(newValue) {
			if err := a.editValue(keyValue, newValue); err != nil {
				return fmt.Errorf("edit value: %w", err)
			}
		}
	}
	return nil
}

func (a *EditMapValueAction) editKey(keyValue *ast.MappingValueNode, newKey any) error {
	comment := keyValue.Key.GetComment()
	v, err := yaml.ValueToNode(newKey)
	if err != nil {
		return err
	}
	if err := v.SetComment(comment); err != nil {
		return fmt.Errorf("set comment to new key: %w", err)
	}
	k, ok := v.(ast.MapKeyNode)
	if !ok {
		return errors.New("failed to convert value to map key node")
	}
	keyValue.Key = k
	return nil
}

func (a *EditMapValueAction) editValue(keyValue *ast.MappingValueNode, newValue any) error {
	comment := keyValue.Value.GetComment()
	v, err := yaml.ValueToNode(newValue)
	if err != nil {
		return err
	}
	if err := v.SetComment(comment); err != nil {
		return fmt.Errorf("set comment to new value: %w", err)
	}
	keyValue.Value = v
	return nil
}
