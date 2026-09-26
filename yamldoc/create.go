package yamldoc

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

// NewNode creates a Node from a Go value like doc.createNode of eemeli/yaml.
// If v is already a Node, it is returned as is.
// Go maps are marshaled in the key order, so use yaml.MapSlice to keep the order of keys.
func NewNode(v any) (Node, error) {
	if n, ok := v.(Node); ok {
		return n, nil
	}
	n, err := valueToAST(v)
	if err != nil {
		return nil, err
	}
	return wrapFresh(n)
}

// ParseNode parses a YAML text and creates a Node.
// Unlike NewNode, comments in the text are kept.
func ParseNode(b []byte) (Node, error) {
	n, err := parseSnippet(b)
	if err != nil {
		return nil, err
	}
	return wrapFresh(n)
}

func wrapFresh(n ast.Node) (Node, error) {
	node, err := wrap(n, nil, true)
	if err != nil {
		return nil, err
	}
	if s, ok := node.(*Scalar); ok {
		s.standalone = true
	}
	return node, nil
}

// NewPair creates a key-value pair like doc.createPair of eemeli/yaml.
// key must be a scalar value.
func NewPair(key, value any) (*Pair, error) {
	k, err := NewNode(key)
	if err != nil {
		return nil, fmt.Errorf("create a key: %w", err)
	}
	ks, ok := k.(*Scalar)
	if !ok {
		return nil, errors.New("key must be a scalar")
	}
	v, err := NewNode(value)
	if err != nil {
		return nil, fmt.Errorf("create a value: %w", err)
	}
	return &Pair{Key: ks, Value: v}, nil
}

// NewMap creates an empty map.
func NewMap() *Map {
	return &Map{}
}

// NewSeq creates an empty sequence.
func NewSeq() *Seq {
	return &Seq{}
}

// initAST creates the AST of a map created by NewMap or &Map{}.
func (m *Map) initAST() error {
	if m.node != nil {
		return nil
	}
	n, err := parseSnippet([]byte("{}"))
	if err != nil {
		return err
	}
	mn, ok := n.(*ast.MappingNode)
	if !ok {
		return fmt.Errorf("unexpected node type: %s", n.Type())
	}
	mn.IsFlowStyle = false
	m.node = mn
	m.outer = mn
	return nil
}

// initAST creates the AST of a sequence created by NewSeq or &Seq{}.
func (s *Seq) initAST() error {
	if s.node != nil {
		return nil
	}
	n, err := parseSnippet([]byte("[]"))
	if err != nil {
		return err
	}
	sn, ok := n.(*ast.SequenceNode)
	if !ok {
		return fmt.Errorf("unexpected node type: %s", n.Type())
	}
	sn.IsFlowStyle = false
	s.node = sn
	s.outer = sn
	s.entries = map[Node]*ast.SequenceEntryNode{}
	return nil
}

// Decode decodes the scalar to a Go value.
func (s *Scalar) Decode(v any) error { return decodeNode(s, v) }

// Decode decodes the map to a Go value.
func (m *Map) Decode(v any) error { return decodeNode(m, v) }

// Decode decodes the sequence to a Go value.
func (s *Seq) Decode(v any) error { return decodeNode(s, v) }

// Decode decodes the alias to a Go value.
// This fails because the anchor can't be resolved from the alias alone.
// Use Document.Decode to decode a document including aliases.
func (a *Alias) Decode(v any) error { return decodeNode(a, v) }

func decodeNode(n Node, v any) error {
	a, err := n.sync(n.base().origParent)
	if err != nil {
		return err
	}
	if err := yaml.NodeToValue(a, v); err != nil {
		return fmt.Errorf("decode a node: %w", err)
	}
	return nil
}

// Decode decodes the document to a Go value.
func (d *Document) Decode(v any) error {
	if err := d.sync(); err != nil {
		return err
	}
	if d.node.Body == nil {
		return nil
	}
	if err := yaml.NodeToValue(d.node.Body, v); err != nil {
		return fmt.Errorf("decode a document: %w", err)
	}
	return nil
}

// GetAs gets the node at the path from n and decodes it to T.
// The second return value is false if the node isn't found.
//
//	v, ok, err := yamldoc.GetAs[string](doc.Contents, "jobs", "test", "runs-on")
func GetAs[T any](n Node, path ...any) (T, bool, error) {
	var v T
	target := getIn(n, path)
	if target == nil {
		return v, false, nil
	}
	if err := target.Decode(&v); err != nil {
		return v, true, err
	}
	return v, true, nil
}
