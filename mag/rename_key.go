package mag

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// RenameKey returns a MapAction renaming given keys from a map.
func RenameKey(key, newKey any) MapAction {
	return &EditMapAction{
		Edit: func(m *MapValue, _ func(any) error) ([]Change, error) {
			kv, ok := m.Map[key]
			if !ok {
				return nil, nil
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

type ChangeRenameKey struct {
	Node *ast.MappingValueNode
	Key  any
}

func (a *ChangeRenameKey) Run() error {
	oldToken := a.Node.Key.GetToken()
	comment := a.Node.Key.GetComment()
	v, err := yaml.ValueToNode(a.Key)
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
