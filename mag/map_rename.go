package mag

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml/ast"
)

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

// RenameKey returns a MapAction renaming given keys from a map.
// whenDuplicate specifies the behavior when the new key already exists.
func RenameKey(key, newKey any, whenDuplicate WhenDuplicateKey) MappingNodeAction {
	return &editMapAction[any, any]{
		Edit: func(m *Map[any, any]) error {
			kv, ok := m.Map[key]
			if !ok {
				return nil
			}
			if key == newKey {
				return nil
			}
			if existing, ok := m.Map[newKey]; ok {
				switch whenDuplicate {
				case Skip:
					return nil
				case IgnoreExistingKey:
					if err := RemoveKeysFromMappingNode(m.Node, kv.Index); err != nil {
						return err
					}
					return SetValueToMappingValue(existing.Node, kv.Value, false)
				case RemoveOldKey:
					return RemoveKeysFromMappingNode(m.Node, kv.Index)
				case RaiseError:
					return fmt.Errorf("key %v already exists", newKey)
				default:
					return fmt.Errorf("unknown duplicate key behavior: %v", whenDuplicate)
				}
			}
			return RenameKeyOfMappingValueNode(kv.Node, newKey)
		},
	}
}

// RenameKeyOfMappingValueNode renames the key of a mapping value node.
func RenameKeyOfMappingValueNode(node *ast.MappingValueNode, newKey any) error {
	oldToken := node.Key.GetToken()
	comment := node.Key.GetComment()
	v, err := valueToNode(newKey)
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
	node.Key = k
	return nil
}
