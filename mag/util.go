package mag

import "reflect"

func unifyInt(value any) any {
	switch v := value.(type) {
	case int64:
		return int(v)
	case uint64:
		return int(v)
	default:
		return value
	}
}

func compareKey(key, keyNodeValue any) bool {
	uKey := unifyInt(key)
	uKeyNodeValue := unifyInt(keyNodeValue)
	return reflect.DeepEqual(uKey, uKeyNodeValue)
}
