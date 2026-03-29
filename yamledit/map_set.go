package yamledit

import (
	"errors"
	"fmt"
	"slices"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

// SetKey returns a MapAction setting given key and value.
// SetKeyOption changes the behavior of SetKey.
// If SetKeyOption is nil, the new key-value pair will be appended to the end of the map, and if the key exists the value will be updated.
func SetKey(key, value any, opt *SetKeyOption) MappingNodeAction {
	return &editMapAction[any, any]{
		Edit: func(m *Map[any, any]) error {
			node, ok := m.Map[key]
			if !ok {
				if opt.GetIgnoreIfKeyNotExist() {
					return nil
				}
				mvn, err := toMappingValueNode(key, value)
				if err != nil {
					return fmt.Errorf("convert key/value to node: %w", err)
				}
				copyColumnFromSibling(mvn, m.Node)
				idx := findInsertIndex(opt.GetInsertLocations(), m.KeyValues)
				m.Node.Values = slices.Insert(m.Node.Values, idx, mvn)
				return nil
			}
			if opt.GetIgnoreIfKeyExist() {
				return nil
			}
			return SetValueToMappingValue(node.Node, value, opt.GetClearComment())
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
	// If true, SetKey will clear the comment of the existing key-value pair if the key already exists.
	ClearComment bool
}

func (o *SetKeyOption) GetIgnoreIfKeyNotExist() bool {
	return o != nil && o.IgnoreIfKeyNotExist
}

func (o *SetKeyOption) GetIgnoreIfKeyExist() bool {
	return o != nil && o.IgnoreIfKeyExist
}

func (o *SetKeyOption) GetClearComment() bool {
	return o != nil && o.ClearComment
}

func (o *SetKeyOption) GetInsertLocations() []*InsertLocation {
	if o == nil {
		return nil
	}
	return o.InsertLocations
}

// SetValueToMappingValue sets the value of a mapping value node.
func SetValueToMappingValue(node *ast.MappingValueNode, value any, clearComment bool) error {
	v, err := valueToNode(value)
	if err != nil {
		return fmt.Errorf("convert value to node: %w", err)
	}
	cmt := node.Value.GetComment()
	if clearComment {
		cmt = nil
	}
	if err := v.SetComment(cmt); err != nil {
		return fmt.Errorf("set comment to new value: %w", err)
	}
	adjustChildColumns(v, node.Key.GetToken().Position.Column)
	node.Value = v
	return nil
}

// adjustChildColumns adjusts column positions of child MappingValueNodes
// when the value is a MappingNode, so they are indented correctly
// relative to their parent key.
func adjustChildColumns(v ast.Node, parentKeyColumn int) {
	mn, ok := v.(*ast.MappingNode)
	if !ok {
		return
	}
	childCol := parentKeyColumn + 2 //nolint:mnd
	for _, child := range mn.Values {
		if tok := child.Key.GetToken(); tok != nil && tok.Position != nil {
			tok.Position.Column = childCol
		}
		if tok := child.GetToken(); tok != nil && tok.Position != nil {
			tok.Position.Column = childCol
		}
		// Recursively adjust nested mappings
		adjustChildColumns(child.Value, childCol)
	}
}

func findInsertIndex(locations []*InsertLocation, kvs []*KeyValue[any]) int {
	for _, loc := range locations {
		if loc.First {
			return 0
		}
		if loc.BeforeKey != nil {
			idx := slices.IndexFunc(kvs, func(v *KeyValue[any]) bool {
				return compareKey(loc.BeforeKey, v.Key)
			})
			if idx != -1 {
				return idx
			}
		}
		if loc.AfterKey != nil {
			idx := slices.IndexFunc(kvs, func(v *KeyValue[any]) bool {
				return compareKey(loc.AfterKey, v.Key)
			})
			if idx != -1 {
				return idx + 1
			}
		}
	}
	return len(kvs)
}

// copyColumnFromSibling copies the column position from an existing sibling
// in the mapping node to the new MappingValueNode so that the new node is
// serialized at the correct indentation level.
func copyColumnFromSibling(mvn *ast.MappingValueNode, parent *ast.MappingNode) {
	if len(parent.Values) == 0 {
		return
	}
	col := parent.Values[0].Key.GetToken().Position.Column
	mvn.Key.GetToken().Position.Column = col
	mvn.GetToken().Position.Column = col
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
