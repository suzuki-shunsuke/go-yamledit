package mag

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml/ast"
)

// RenameKey returns a MapAction renaming given keys from a map.
// whenDuplicate specifies the behavior when the new key already exists.
func RenameKey(key, newKey any, whenDuplicate WhenDuplicateKey) MapAction {
	return &EditMapAction[any, any]{
		Edit: func(m *MapValue[any, any]) ([]Change, error) {
			kv, ok := m.Map[key]
			if !ok {
				return nil, nil
			}
			if key == newKey {
				return nil, nil
			}
			if existing, ok := m.Map[newKey]; ok {
				switch whenDuplicate {
				case Skip:
					return nil, nil
				case IgnoreExistingKey:
					return []Change{
						&ChangeRemoveKeys{
							Indexes: []int{kv.Index},
							Node:    m.Node,
						},
						&ChangeSetValue{
							Node:  existing.Node,
							Value: kv.Value,
						},
					}, nil
				case RemoveOldKey:
					return []Change{
						&ChangeRemoveKeys{
							Indexes: []int{kv.Index},
							Node:    m.Node,
						},
					}, nil
				case RaiseError:
					return nil, fmt.Errorf("key %v already exists", newKey)
				default:
					return nil, fmt.Errorf("unknown duplicate key behavior: %v", whenDuplicate)
				}
			}
			return []Change{
				&ChangeRenameKey{
					Key:  newKey,
					Node: kv.Node,
				},
			}, nil
		},
	}
}

// WhenDuplicateKey specifies the behavior when the new key already exists.
type WhenDuplicateKey int

const (
	// Skip doesn't change the node.
	Skip WhenDuplicateKey = iota
	// IgnoreExistingKey overwrites the existing key and value.
	IgnoreExistingKey
	// RemoveOldKey removes the old key and value.
	RemoveOldKey
	// RaiseError returns an error when the new key already exists.
	RaiseError
)

type ChangeRenameKey struct {
	Node *ast.MappingValueNode
	Key  any
}

func (a *ChangeRenameKey) Run() error {
	oldToken := a.Node.Key.GetToken()
	comment := a.Node.Key.GetComment()
	v, err := valueToNode(a.Key)
	if err != nil {
		return err
	}
	// Preserve the original token's position so that indentation is maintained
	// when the AST is serialized back to a string.
	v.GetToken().Position = oldToken.Position
	if err := v.SetComment(comment); err != nil {
		return fmt.Errorf("set comment to new key: %w", err)
	}
	k, ok := v.(ast.MapKeyNode)
	if !ok {
		return errors.New("failed to convert value to map key node")
	}
	a.Node.Key = k
	return nil
}
