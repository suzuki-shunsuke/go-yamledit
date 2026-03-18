package mag

import (
	"errors"
	"fmt"

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
	// Add is a function that returns the key, value, and index to insert into the map.
	Add AddMapKey
}

// AddMapKey is a function returning the key, value, and index to insert into a map.
// If error is ErrNoop, no item will be added.
// If the index is negative, the key will be inserted at the end.
type AddMapKey func(node *ast.MappingNode) (any, any, int, error)

// Run adds a key to the given map.
func (a *AddMapKeyAction) Run(m *ast.MappingNode) error {
	if a.Add == nil {
		return errors.New("Add is not set")
	}
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

// AddToMap returns a MapAction adding the given key and value at the given index.
// If the index is negative, the key will be inserted at the end.
func AddToMap(key, value any, idx int) MapAction {
	s := &staticAddMapKeyEditor{
		key:   key,
		value: value,
		idx:   idx,
	}
	return &AddMapKeyAction{
		Add: s.Add,
	}
}
