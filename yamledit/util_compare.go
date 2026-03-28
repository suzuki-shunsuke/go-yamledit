package yamledit

import (
	"fmt"
	"reflect"
)

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
