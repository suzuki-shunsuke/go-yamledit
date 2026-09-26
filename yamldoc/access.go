package yamldoc

import (
	"errors"
	"fmt"
	"slices"
)

// IndexOf returns the index of the pair with the key in Items.
// It returns -1 if the key isn't found.
// This is useful to insert a pair at a specific position.
//
//	m.Items = slices.Insert(m.Items, m.IndexOf("name")+1, pair)
func (m *Map) IndexOf(key any) int {
	return slices.IndexFunc(m.Items, func(p *Pair) bool {
		return p != nil && p.Key != nil && valueEqual(p.Key.Value, key)
	})
}

// Pair returns the pair with the key.
// It returns nil if the key isn't found.
func (m *Map) Pair(key any) *Pair {
	if i := m.IndexOf(key); i >= 0 {
		return m.Items[i]
	}
	return nil
}

// Get returns the value of the key.
// It returns nil if the key isn't found.
func (m *Map) Get(key any) Node {
	if p := m.Pair(key); p != nil {
		return p.Value
	}
	return nil
}

// Has returns true if the map has the key.
func (m *Map) Has(key any) bool {
	return m.IndexOf(key) >= 0
}

// Set sets the value of the key.
// If the key doesn't exist, a new pair is appended to the end of the map.
// value is a Node or a Go value.
func (m *Map) Set(key, value any) error {
	if p := m.Pair(key); p != nil {
		return p.SetValue(value)
	}
	p, err := NewPair(key, value)
	if err != nil {
		return err
	}
	m.Items = append(m.Items, p)
	return nil
}

// Delete deletes the pair with the key.
// It returns true if the key is found.
func (m *Map) Delete(key any) bool {
	i := m.IndexOf(key)
	if i < 0 {
		return false
	}
	m.Items = slices.Delete(m.Items, i, i+1)
	return true
}

// Keys returns the keys of the map.
func (m *Map) Keys() []any {
	keys := make([]any, 0, len(m.Items))
	for _, p := range m.Items {
		if p != nil && p.Key != nil {
			keys = append(keys, p.Key.Value)
		}
	}
	return keys
}

// SetValue sets the value of the pair.
// If both the current value and the new value are scalars, only the value is changed and the comment is kept.
func (p *Pair) SetValue(value any) error {
	n, err := NewNode(value)
	if err != nil {
		return err
	}
	if cur, ok := p.Value.(*Scalar); ok {
		if s, ok := n.(*Scalar); ok && s.node != nil {
			cur.Value = s.Value
			return nil
		}
	}
	p.Value = n
	return nil
}

// Get returns the item at the index.
// It returns nil if the index is out of range.
func (s *Seq) Get(index int) Node {
	if index < 0 || index >= len(s.Items) {
		return nil
	}
	return s.Items[index]
}

// Has returns true if the index is in range.
func (s *Seq) Has(index int) bool {
	return index >= 0 && index < len(s.Items)
}

// Set sets the item at the index.
// If both the current item and the new item are scalars, only the value is changed and the comment is kept.
func (s *Seq) Set(index int, value any) error {
	if !s.Has(index) {
		return fmt.Errorf("index out of range: %d", index)
	}
	n, err := NewNode(value)
	if err != nil {
		return err
	}
	if cur, ok := s.Items[index].(*Scalar); ok {
		if sc, ok := n.(*Scalar); ok && sc.node != nil {
			cur.Value = sc.Value
			return nil
		}
	}
	s.Items[index] = n
	return nil
}

// Delete deletes the item at the index.
// It returns true if the index is in range.
func (s *Seq) Delete(index int) bool {
	if !s.Has(index) {
		return false
	}
	s.Items = slices.Delete(s.Items, index, index+1)
	return true
}

// Add appends a value to the end of the sequence.
func (s *Seq) Add(value any) error {
	n, err := NewNode(value)
	if err != nil {
		return err
	}
	s.Items = append(s.Items, n)
	return nil
}

func seqIndex(key any) (int, bool) {
	i, ok, big := toInt(key)
	if !ok || big {
		return 0, false
	}
	return int(i), true
}

func getChild(n Node, key any) Node {
	switch v := n.(type) {
	case *Map:
		return v.Get(key)
	case *Seq:
		i, ok := seqIndex(key)
		if !ok {
			return nil
		}
		return v.Get(i)
	default:
		return nil
	}
}

func getIn(n Node, path []any) Node {
	for _, key := range path {
		if n == nil {
			return nil
		}
		n = getChild(n, key)
	}
	return n
}

func setChild(n Node, key, value any) error {
	switch v := n.(type) {
	case *Map:
		return v.Set(key, value)
	case *Seq:
		i, ok := seqIndex(key)
		if !ok {
			return fmt.Errorf("sequence index must be an integer: %v", key)
		}
		if i == len(v.Items) {
			return v.Add(value)
		}
		return v.Set(i, value)
	default:
		return fmt.Errorf("can't set a value to %T", n)
	}
}

// setIn sets the value at the path.
// Missing intermediate maps are created.
func setIn(n Node, path []any, value any) error {
	if len(path) == 0 {
		return errors.New("path is empty")
	}
	for i, key := range path[:len(path)-1] {
		child := getChild(n, key)
		if child == nil {
			m := NewMap()
			if err := setChild(n, key, m); err != nil {
				return fmt.Errorf("path %v: %w", path[:i+1], err)
			}
			child = m
		}
		n = child
	}
	return setChild(n, path[len(path)-1], value)
}

func deleteIn(n Node, path []any) bool {
	if len(path) == 0 {
		return false
	}
	switch v := getIn(n, path[:len(path)-1]).(type) {
	case *Map:
		return v.Delete(path[len(path)-1])
	case *Seq:
		i, ok := seqIndex(path[len(path)-1])
		return ok && v.Delete(i)
	default:
		return false
	}
}

// GetIn returns the node at the path.
// Keys of maps and indexes of sequences are mixed in the path.
// It returns nil if the node isn't found.
func (m *Map) GetIn(path ...any) Node { return getIn(m, path) }

// HasIn returns true if the node at the path exists.
func (m *Map) HasIn(path ...any) bool { return getIn(m, path) != nil }

// SetIn sets the value at the path.
// Missing intermediate maps are created.
func (m *Map) SetIn(path []any, value any) error { return setIn(m, path, value) }

// DeleteIn deletes the node at the path.
// It returns true if the node is found.
func (m *Map) DeleteIn(path ...any) bool { return deleteIn(m, path) }

// GetIn returns the node at the path.
func (s *Seq) GetIn(path ...any) Node { return getIn(s, path) }

// HasIn returns true if the node at the path exists.
func (s *Seq) HasIn(path ...any) bool { return getIn(s, path) != nil }

// SetIn sets the value at the path.
func (s *Seq) SetIn(path []any, value any) error { return setIn(s, path, value) }

// DeleteIn deletes the node at the path.
func (s *Seq) DeleteIn(path ...any) bool { return deleteIn(s, path) }

// Get returns the value of the key of the root map, or the item of the root sequence.
func (d *Document) Get(key any) Node { return getIn(d.Contents, []any{key}) }

// Has returns true if the root collection has the key.
func (d *Document) Has(key any) bool { return d.Get(key) != nil }

// Set sets the value of the key of the root collection.
// If the document is empty, a map is created.
func (d *Document) Set(key, value any) error { return d.SetIn([]any{key}, value) }

// Delete deletes the key of the root collection.
func (d *Document) Delete(key any) bool { return deleteIn(d.Contents, []any{key}) }

// GetIn returns the node at the path.
func (d *Document) GetIn(path ...any) Node { return getIn(d.Contents, path) }

// HasIn returns true if the node at the path exists.
func (d *Document) HasIn(path ...any) bool { return getIn(d.Contents, path) != nil }

// SetIn sets the value at the path.
// If the document is empty, a map is created.
func (d *Document) SetIn(path []any, value any) error {
	if d.Contents == nil {
		d.Contents = NewMap()
	}
	return setIn(d.Contents, path, value)
}

// DeleteIn deletes the node at the path.
func (d *Document) DeleteIn(path ...any) bool { return deleteIn(d.Contents, path) }
