package mag

import (
	"errors"
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
		if err := a.removeKeyFromMap(e); err != nil {
			return err
		}
	}
	return nil
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
