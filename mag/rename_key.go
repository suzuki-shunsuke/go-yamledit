package mag

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

type RenameKeyAction struct {
	YAMLPath string
	OldKey   any
	NewKey   any
}

func (a *RenameKeyAction) Run(node ast.Node) error {
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
		return a.renameMapKey(v)
	case *ast.SequenceNode:
		for _, elem := range v.Values {
			m, ok := elem.(*ast.MappingNode)
			if !ok {
				continue
			}
			if err := a.renameMapKey(m); err != nil {
				return err
			}
		}
		return nil
	default:
		return nil
	}
}

func (a *RenameKeyAction) renameMapKey(m *ast.MappingNode) error {
	mapIter := m.MapRange()
	for mapIter.Next() {
		kv := mapIter.KeyValue()
		var keyV any
		if err := yaml.NodeToValue(kv.Key, &keyV); err != nil {
			return err
		}
		if !compareKey(a.OldKey, keyV) {
			continue
		}
		comment := kv.Key.GetComment()
		v, err := yaml.ValueToNode(a.NewKey)
		if err != nil {
			return err
		}
		v.SetComment(comment)
		k, ok := v.(ast.MapKeyNode)
		if !ok {
			return errors.New("failed to convert value to map key node")
		}
		kv.Key = k
		return nil
	}
	return nil
}
