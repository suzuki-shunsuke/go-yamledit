package mag

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

type noop struct{}

// NoChange is a sentinel value that indicates no change should be made against mapping key or value.
var NoChange = noop{} //nolint:gochecknoglobals

func isChanged(value any) bool {
	_, ok := value.(noop)
	return !ok
}

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
