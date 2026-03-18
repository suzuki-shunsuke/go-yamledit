package mag

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// Map returns Action editing maps.
func Map(yamlPath string, actions ...MapAction) Action {
	return &MapActions{
		YAMLPath: yamlPath,
		Actions:  actions,
	}
}

// MapAction represents an action editing a map.
type MapAction interface {
	Run(m *ast.MappingNode) error
}

// MapActions is an Action editing maps.
type MapActions struct {
	// YAMLPath is a path to edited maps.
	// e.g. "$.reviewer"
	// https://github.com/goccy/go-yaml/blob/v1.19.2/path.go#L17-L22
	YAMLPath string
	Actions  []MapAction
}

// Run edits maps.
func (a *MapActions) Run(node ast.Node) error {
	if a.YAMLPath == "" {
		return errors.New("YAMLPath is not set")
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
		m, ok := elem.(*ast.MappingNode)
		if !ok {
			return fmt.Errorf("expected a mapping node, got %s", elem.Type().String())
		}
		if err := a.run(m); err != nil {
			return err
		}
	}
	return nil
}

func (a *MapActions) run(m *ast.MappingNode) error {
	for _, action := range a.Actions {
		if err := action.Run(m); err != nil {
			return err
		}
	}
	return nil
}
