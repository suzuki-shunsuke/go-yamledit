package mag

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

// TODO
// map
//   [x] Add key to a map
//   [ ] Sort keys
// comment
//   [ ] Add comment
//   [ ] Remove comment
//   [ ] Edit comment

// AddMapKeyAction represents an action to add a key to a map.
type AddMapKeyAction struct {
	// YAMLPath is a path to YAML mapping nodes that new key will be added to.
	// e.g. "$.reviewer"
	// https://github.com/goccy/go-yaml/blob/v1.19.2/path.go#L17-L22
	YAMLPath string
	// Add is a function that returns the key, value, and index to insert into the map.
	Add AddMapKey
}

// AddMapKey is a function returning the key, value, and index to insert into a map.
// If error is ErrNoop, no item will be added.
// If the index is negative, the key will be inserted at the end.
type AddMapKey func(node *ast.MappingNode) (any, any, int, error)

// Run adds a pair of key and value to a YAML mapping node.
func (a *AddMapKeyAction) Run(node ast.Node) error {
	if a.Add == nil {
		return errors.New("add is not set")
	}
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
		if err := a.add(m); err != nil {
			return err
		}
	}
	return nil
}

func (a *AddMapKeyAction) add(m *ast.MappingNode) error {
	k, v, idx, err := a.Add(m)
	if errors.Is(err, ErrNoop) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("add a key to map: %w", err)
	}
	if idx < 0 {
		idx += len(m.Values)
	}

	kn, err := valueToNode(k)
	if err != nil {
		return fmt.Errorf("convert key to node: %w", err)
	}
	keyNode, ok := kn.(ast.MapKeyNode)
	if !ok {
		return errors.New("key is not a valid map key type")
	}

	vn, err := valueToNode(v)
	if err != nil {
		return fmt.Errorf("convert value to node: %w", err)
	}

	tk := token.MappingValue(&token.Position{})
	mvn := ast.MappingValue(tk, keyNode, vn)
	m.Values = append(m.Values[:idx], append([]*ast.MappingValueNode{mvn}, m.Values[idx:]...)...)
	return nil
}

type staticAddMapKeyEditor struct {
	key   any
	value any
	idx   int
}

func (e *staticAddMapKeyEditor) Add(_ *ast.MappingNode) (any, any, int, error) {
	return e.key, e.value, e.idx, nil
}

// NewStaticAddMapKeyEditor returns an AddMapKey function adding the given key and value, to the given index.
func NewStaticAddMapKeyEditor(key, value any, idx int) AddMapKey {
	s := &staticAddMapKeyEditor{
		key:   key,
		value: value,
		idx:   idx,
	}
	return s.Add
}
