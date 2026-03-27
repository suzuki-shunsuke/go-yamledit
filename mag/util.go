package mag

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

// BytesToNode parses the given YAML bytes and returns the root node.
// Returns an error if the bytes cannot be parsed.
// YAML should be a single document.
func BytesToNode(b []byte) (ast.Node, error) {
	file, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return file.Docs[0].Body, nil
}

// YAML converts yaml bytes to a YAMLBytes struct.
// This is useful to pass YAML strings to functions without temporary variables and error handling.
func YAML(b []byte) *YAMLBytes {
	return &YAMLBytes{b: b}
}

// YAMLBytes holds a YAML document.
// This is converted to ast.Node internally.
type YAMLBytes struct {
	b []byte
}

type noop struct{}

// NoChange is a sentinel value that indicates no change should be made against mapping key or value.
var NoChange = noop{} //nolint:gochecknoglobals

func unifyInt(value any) (any, bool) {
	switch v := value.(type) {
	case int, int64, uint64:
		return fmt.Sprintf("%d", v), true
	default:
		return value, false
	}
}

func compareKey(key, keyNodeValue any) bool {
	uKey, b1 := unifyInt(key)
	uKeyNodeValue, b2 := unifyInt(keyNodeValue)
	return b1 == b2 && reflect.DeepEqual(uKey, uKeyNodeValue)
}

func flatten(node ast.Node, depth int) ([]ast.Node, error) {
	if depth == 0 {
		return []ast.Node{node}, nil
	}

	seq, ok := node.(*ast.SequenceNode)
	if !ok {
		if depth == -1 {
			return []ast.Node{node}, nil
		}
		return nil, fmt.Errorf("expected a sequence node: %s", node.Type().String())
	}
	ret := []ast.Node{}
	for _, elem := range seq.Values {
		n := depth - 1
		if depth == -1 {
			n = -1
		}
		nodes, err := flatten(elem, n)
		if err != nil {
			return nil, err
		}
		ret = append(ret, nodes...)
	}
	return ret, nil
}

func getDepthByPath(yamlPath string) int { //nolint:cyclop
	count := 0
	inQuote := false
	for i := 0; i < len(yamlPath); i++ {
		ch := yamlPath[i]
		if ch == '\\' && inQuote && i+1 < len(yamlPath) && yamlPath[i+1] == '\'' {
			i++ // skip escaped quote
			continue
		}
		if ch == '\'' {
			inQuote = !inQuote
			continue
		}
		if inQuote {
			continue
		}
		if i+3 <= len(yamlPath) && yamlPath[i:i+3] == "[*]" {
			count++
			i += 2 // skip rest of [*]
			continue
		}
		if ch == '.' && i+1 < len(yamlPath) && yamlPath[i+1] == '.' {
			count++
			i++ // skip second dot
			continue
		}
	}
	return count
}

func valueToNode(value any) (ast.Node, error) {
	if node, ok := value.(ast.Node); ok {
		return node, nil
	}
	if b, ok := value.(*YAMLBytes); ok {
		return BytesToNode(b.b)
	}
	valWithComment := toValueWithComment(value)
	v, err := yaml.ValueToNode(valWithComment.Value)
	if err != nil {
		return nil, err
	}
	if valWithComment.Comment == "" {
		return v, nil
	}
	if err := v.SetComment(commentGroupFromString(valWithComment.Comment)); err != nil {
		return nil, err
	}
	return v, nil
}

func normalizeIndexes(indexes []int, size int) error {
	for i, idx := range indexes {
		newIdx, err := checkIndex(idx, size)
		if err != nil {
			return err
		}
		indexes[i] = newIdx
	}
	return nil
}

// checkInsertIndex normalizes an index for insertion into a list.
// idx == size means append to the end. Negative indexes count from
// the end, where -1 means append after the last element.
func checkInsertIndex(idx, size int) (int, error) {
	if idx > size {
		return 0, errors.New("index is larger than the size of the list")
	}
	if idx >= 0 {
		return idx, nil
	}
	newIdx := size + idx + 1
	if newIdx < 0 {
		return 0, errors.New("the negative index is smaller than the size of the list")
	}
	return newIdx, nil
}

// checkIndex normalizes an index for accessing an existing element in a list.
// Unlike checkInsertIndex, idx must be strictly less than size.
// Negative indexes count from the end, where -1 means the last element.
func checkIndex(idx, size int) (int, error) {
	if idx >= size {
		return 0, errors.New("index is larger than the size of the list")
	}
	if idx >= 0 {
		return idx, nil
	}
	newIdx := size + idx
	if newIdx < 0 {
		return 0, errors.New("the negative index is smaller than the size of the list")
	}
	return newIdx, nil
}
