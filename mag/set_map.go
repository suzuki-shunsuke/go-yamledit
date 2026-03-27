package mag

import (
	"errors"
	"fmt"
	"slices"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

// TODO
// comment
//   [ ] Add comment
//   [ ] Remove comment
//   [ ] Edit comment

// SetKey returns a MapAction setting given key and value.
// SetKeyOption changes the behavior of SetKey.
// If SetKeyOption is nil, the new key-value pair will be appended to the end of the map, and if the key exists the value will be updated.
func SetKey(key, value any, opt *SetKeyOption) MapAction {
	return &EditMapAction[any, any]{
		Edit: func(m *MapValue[any, any]) ([]Change, error) {
			node, ok := m.Map[key]
			if !ok {
				if opt.GetIgnoreIfKeyNotExist() {
					return nil, nil
				}
				mvn, err := toMappingValueNode(key, value)
				if err != nil {
					return nil, fmt.Errorf("convert key/value to node: %w", err)
				}
				changeAddKey := &ChangeAddKey{
					M:     m.Node,
					Nodes: []*ast.MappingValueNode{mvn},
				}
				changes := []Change{changeAddKey}
				for _, location := range opt.GetInsertLocations() {
					if location.First {
						return changes, nil
					}
					if location.BeforeKey != nil {
						idx := slices.IndexFunc(m.KeyValues, func(v *KeyValue[any]) bool {
							return compareKey(location.BeforeKey, v.Key)
						})
						if idx == -1 {
							continue
						}
						changeAddKey.Index = idx
						return changes, nil
					}
					if location.AfterKey != nil {
						idx := slices.IndexFunc(m.KeyValues, func(v *KeyValue[any]) bool {
							return compareKey(location.AfterKey, v.Key)
						})
						if idx == -1 {
							continue
						}
						changeAddKey.Index = idx + 1
						return changes, nil
					}
				}
				changeAddKey.Index = len(m.KeyValues)
				return changes, nil
			}
			if opt.GetIgnoreIfKeyExist() {
				return nil, nil
			}
			return []Change{
				&ChangeSetValue{
					Node:  node.Node,
					Value: value,
				},
			}, nil
		},
	}
}

// InsertLocation specifies the location to insert the new key-value pair.
type InsertLocation struct {
	// If true, the new key-value pair will be inserted at the beginning of the map.
	First bool
	// If not nil, the new key-value pair will be inserted before the key with the given value.
	// If the key is not found, this is ignored.
	BeforeKey any
	// If not nil, the new key-value pair will be inserted after the key with the given value.
	// If the key is not found, this is ignored.
	AfterKey any
}

// SetKeyOption changes the behavior of SetKey.
// By default, the new key-value pair will be appended to the end of the map, and if the key exists the value will be updated.
type SetKeyOption struct {
	// If true, SetKey will not add a new key if the map doesn't have the key.
	IgnoreIfKeyNotExist bool
	// If true, SetKey will not set a new value if the key already exists.
	IgnoreIfKeyExist bool
	// InsertLocations specifies the locations to insert the new key-value pair.
	// The first location that matches the condition will be used.
	// If empty or no location matches the condition, the new key-value pair will be appended to the end of the map.
	InsertLocations []*InsertLocation
}

func (o *SetKeyOption) GetIgnoreIfKeyNotExist() bool {
	return o != nil && o.IgnoreIfKeyNotExist
}

func (o *SetKeyOption) GetIgnoreIfKeyExist() bool {
	return o != nil && o.IgnoreIfKeyExist
}

func (o *SetKeyOption) GetInsertLocations() []*InsertLocation {
	if o == nil {
		return nil
	}
	return o.InsertLocations
}

type ChangeAddKey struct {
	M     *ast.MappingNode
	Nodes []*ast.MappingValueNode
	Index int
}

func (a *ChangeAddKey) Run() error {
	a.M.Values = slices.Insert(a.M.Values, a.Index, a.Nodes...)
	return nil
}

type ChangeSetValue struct {
	Node  *ast.MappingValueNode
	Value any
}

func (a *ChangeSetValue) Run() error {
	v, err := yaml.ValueToNode(a.Value)
	if err != nil {
		return fmt.Errorf("convert value to node: %w", err)
	}
	if err := v.SetComment(a.Node.Value.GetComment()); err != nil {
		return fmt.Errorf("set comment to new value: %w", err)
	}
	a.Node.Value = v
	return nil
}

func toMappingValueNode(k, v any) (*ast.MappingValueNode, error) {
	kn, err := valueToNode(k)
	if err != nil {
		return nil, fmt.Errorf("convert key to node: %w", err)
	}
	keyNode, ok := kn.(ast.MapKeyNode)
	if !ok {
		return nil, errors.New("key is not a valid map key type")
	}

	vn, err := valueToNode(v)
	if err != nil {
		return nil, fmt.Errorf("convert value to node: %w", err)
	}

	return ast.MappingValue(
		token.MappingValue(&token.Position{}),
		keyNode, vn,
	), nil
}
