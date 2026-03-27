package mag

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// EditMapValueAction represents an action to edit a map key and value.
type EditMapValueAction struct {
	// Match filters mapping keys and values to be edited.
	Match MatchMappingValue
	// Edit edits keys and values.
	Edit EditMappingValue
}

// Run edits keys and values of a given map.
func (a *EditMapValueAction) Run(m *ast.MappingNode) error {
	mapIter := m.MapRange()
	for mapIter.Next() {
		keyValue := mapIter.KeyValue()

		f, err := a.Match(keyValue)
		if err != nil {
			return err
		}
		if !f {
			continue
		}

		newKey, newValue, err := a.Edit(keyValue)
		if err != nil {
			return err
		}
		if isChanged(newKey) {
			if err := a.editKey(keyValue, newKey); err != nil {
				return fmt.Errorf("edit key: %w", err)
			}
		}
		if isChanged(newValue) {
			if err := a.editValue(keyValue, newValue); err != nil {
				return fmt.Errorf("edit value: %w", err)
			}
		}
	}
	return nil
}

func (a *EditMapValueAction) editKey(keyValue *ast.MappingValueNode, newKey any) error {
	oldToken := keyValue.Key.GetToken()
	comment := keyValue.Key.GetComment()
	v, err := yaml.ValueToNode(newKey)
	if err != nil {
		return err
	}
	// Preserve the original token's position so that indentation is maintained
	// when the AST is serialized back to a string.
	newToken := v.GetToken()
	newToken.Position = oldToken.Position
	if err := v.SetComment(comment); err != nil {
		return fmt.Errorf("set comment to new key: %w", err)
	}
	k, ok := v.(ast.MapKeyNode)
	if !ok {
		return errors.New("failed to convert value to map key node")
	}
	keyValue.Key = k
	return nil
}

func (a *EditMapValueAction) editValue(keyValue *ast.MappingValueNode, newValue any) error {
	comment := keyValue.Value.GetComment()
	v, err := yaml.ValueToNode(newValue)
	if err != nil {
		return err
	}
	if err := v.SetComment(comment); err != nil {
		return fmt.Errorf("set comment to new value: %w", err)
	}
	keyValue.Value = v
	return nil
}
