package mag

import (
	"fmt"
	"reflect"

	"github.com/goccy/go-yaml/ast"
)

type noop struct{}

var Noop = noop{}

func IsChanged(value any) bool {
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
