package yamldoc

import (
	"math"
	"reflect"
)

// normalizeValue converts integers to int if possible.
// goccy/go-yaml decodes integers to uint64 or int64, which is inconvenient to compare with Go literals.
func normalizeValue(v any) any {
	switch n := v.(type) {
	case uint64:
		if n <= math.MaxInt {
			return int(n)
		}
	case int64:
		if n >= math.MinInt && n <= math.MaxInt {
			return int(n)
		}
	}
	return v
}

// toInt converts a Go integer to int64 or uint64.
// The second value is true if the value is an integer.
// The third value is true if the value is greater than math.MaxInt64.
func toInt(v any) (int64, bool, bool) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() { //nolint:exhaustive
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), true, false
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u := rv.Uint()
		if u > math.MaxInt64 {
			return 0, true, true
		}
		return int64(u), true, false
	default:
		return 0, false, false
	}
}

// valueEqual compares two scalar values.
// Integers are equal if they have the same value regardless of their types.
func valueEqual(a, b any) bool {
	ai, aok, aBig := toInt(a)
	bi, bok, bBig := toInt(b)
	if aok && bok && !aBig && !bBig {
		return ai == bi
	}
	return reflect.DeepEqual(a, b)
}
