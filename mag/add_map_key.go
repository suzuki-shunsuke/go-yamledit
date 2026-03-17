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

type AddMapKeyAction struct {
	YAMLPath string
	Add      AddMapKey
}

type AddMapKey func(node *ast.MappingNode) (any, any, int, error)

func (a *AddMapKeyAction) Run(node ast.Node) error {
	if a.Add == nil {
		return errors.New("add is not set")
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

var NoopError = errors.New("")

func (a *AddMapKeyAction) add(m *ast.MappingNode) error {
	k, v, idx, err := a.Add(m)
	if errors.Is(err, NoopError) {
		return nil
	}
	if idx < 0 {
		idx += len(m.Values)
	}

	kwc := toValueWithComment(k)
	vwc := toValueWithComment(v)

	kn, err := yaml.ValueToNode(kwc.Value)
	if err != nil {
		return fmt.Errorf("convert key to node: %w", err)
	}
	if kwc.Comment != "" {
		if err := kn.SetComment(commentGroupFromString(kwc.Comment)); err != nil {
			return err
		}
	}

	keyNode, ok := kn.(ast.MapKeyNode)
	if !ok {
		return errors.New("key is not a valid map key type")
	}

	vn, err := yaml.ValueToNode(vwc.Value)
	if err != nil {
		return fmt.Errorf("convert value to node: %w", err)
	}
	if vwc.Comment != "" {
		if err := vn.SetComment(commentGroupFromString(vwc.Comment)); err != nil {
			return err
		}
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

func (e *staticAddMapKeyEditor) Add(node *ast.MappingNode) (any, any, int, error) {
	return e.key, e.value, e.idx, nil
}

func NewStaticAddMapKeyEditor(key, value any, idx int) AddMapKey {
	s := &staticAddMapKeyEditor{
		key:   key,
		value: value,
		idx:   idx,
	}
	return s.Add
}
