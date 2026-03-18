package mag

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// Action represents an operation updating YAML AST nodes.
type Action interface {
	// Run modifies the given YAML AST node.
	Run(node ast.Node) error
}

type ListActions struct {
	YAMLPath string
	Actions  []ListAction
}

type ListAction interface {
	Run(seq *ast.SequenceNode) error
}

func (a *ListActions) Run(node ast.Node) error {
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
	nodes, err := flatten(n, getDepthByPath(a.YAMLPath))
	if err != nil {
		return err
	}
	for _, elem := range nodes {
		if err := a.run(elem); err != nil {
			return err
		}
	}
	return nil
}

func (a *ListActions) run(node ast.Node) error {
	seq, ok := node.(*ast.SequenceNode)
	if !ok {
		return fmt.Errorf("expected a sequence node: %s", node.Type().String())
	}
	for _, action := range a.Actions {
		if err := action.Run(seq); err != nil {
			return err
		}
	}
	return nil
}

type MapActions struct {
	YAMLPath string
	Actions  []MapAction
}

type MapAction interface {
	Run(m *ast.MappingNode) error
}

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
