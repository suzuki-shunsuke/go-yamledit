package mag

import "errors"

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
