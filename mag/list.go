package mag

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// List returns Action editing lists.
func List(yamlPath string, actions ...ListAction) Action {
	return &ListActions{
		YAMLPath: yamlPath,
		Actions:  actions,
	}
}

// ListAction represents an action editing a list.
type ListAction interface {
	Run(seq *ast.SequenceNode) error
}

// ListActions is an Action editing lists.
type ListActions struct {
	// YAMLPath is a path to edited lists.
	// e.g. "$.reviewers"
	// https://github.com/goccy/go-yaml/blob/v1.19.2/path.go#L17-L22
	YAMLPath string
	Actions  []ListAction
}

// Run edits lists.
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
