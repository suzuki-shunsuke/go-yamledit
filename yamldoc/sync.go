package yamldoc

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/goccy/go-yaml/token"
)

// sync writes the changes of the tree back to the AST.
// Only the structure is updated here. The indentation is fixed by layout.

func (s *Scalar) sync(parent any) (ast.Node, error) {
	if s.needsRebuild(parent) {
		if err := s.rebuild(); err != nil {
			return nil, err
		}
	}
	if inFlow(parent) && isBlockMultiline(s.node) {
		// A block scalar can't be in a flow collection.
		s.quote()
	}
	if err := s.syncComment(s.node); err != nil {
		return nil, err
	}
	return s.outer, nil
}

// needsRebuild returns true if the AST node must be regenerated from the value.
func (s *Scalar) needsRebuild(parent any) bool {
	if s.node == nil || !valueEqual(s.Value, s.orig) {
		return true
	}
	// A multi-line scalar has its indentation in the original text,
	// so it's regenerated when it's moved to another place.
	return s.origParent != nil && !s.kept(parent) && !s.standalone && isMultiline(s.node)
}

// syncComment sets the comment to the AST node if it's changed.
func (b *nodeBase) syncComment(n ast.Node) error {
	if b.comment == b.origComment {
		return nil
	}
	if err := n.SetComment(newCommentGroup(b.comment)); err != nil {
		return fmt.Errorf("set a comment: %w", err)
	}
	b.origComment = b.comment
	return nil
}

// rebuild regenerates the AST node from the value.
func (s *Scalar) rebuild() error {
	n, err := valueToAST(s.Value)
	if err != nil {
		return err
	}
	if _, ok := n.(ast.ScalarNode); !ok {
		return fmt.Errorf("Scalar.Value must be a scalar value, but got %T. Use Map.Set or Seq.Set to set a collection", s.Value)
	}
	s.replace(n)
	s.orig = s.Value
	return nil
}

// quote converts a multi-line string to a double-quoted string.
func (s *Scalar) quote() {
	v := s.String()
	s.replace(ast.String(token.DoubleQuote(v, strconv.Quote(v), &token.Position{Column: 1})))
}

// replace replaces the AST node keeping the comment.
func (s *Scalar) replace(n ast.Node) {
	if s.node != nil {
		n.SetComment(s.node.GetComment()) //nolint:errcheck
	}
	s.node = n
	s.setInner(n)
	s.standalone = true
}

// isBlockMultiline returns true if the node is a block scalar or a multi-line plain scalar.
func isBlockMultiline(n ast.Node) bool {
	if n == nil {
		return false
	}
	if _, ok := n.(*ast.LiteralNode); ok {
		return true
	}
	tk := n.GetToken()
	return tk.Type != token.DoubleQuoteType && tk.Type != token.SingleQuoteType && strings.Contains(tk.Value, "\n")
}

// inFlow returns true if the parent is a flow collection.
func inFlow(parent any) bool {
	switch p := parent.(type) {
	case *Pair:
		return p.inFlow
	case *Seq:
		return p.node != nil && p.node.IsFlowStyle
	default:
		return false
	}
}

// fixFlowStyle updates the style of a collection before its items are synced.
// A collection in a flow collection must be flow style.
// An empty flow collection outside flow collections becomes block style when items are added.
//
//	foo: {} => foo:
//	             bar: 1
func fixFlowStyle(b *nodeBase, flow *bool, parent any, empty bool) {
	switch {
	case inFlow(parent):
		*flow = true
	case *flow && empty:
		*flow = false
		// Lay it out again because the column of "{" or "[" isn't the column of items.
		b.origParent = nil
	}
}

// isMultiline returns true if the node is a multi-line scalar.
func isMultiline(n ast.Node) bool {
	if n == nil {
		return false
	}
	if _, ok := n.(*ast.LiteralNode); ok {
		return true
	}
	return strings.Contains(n.GetToken().Value, "\n")
}

func (m *Map) sync(parent any) (ast.Node, error) {
	if err := m.initAST(); err != nil {
		return nil, err
	}
	if len(m.Items) > 0 {
		fixFlowStyle(&m.nodeBase, &m.node.IsFlowStyle, parent, len(m.node.Values) == 0)
	}
	values := make([]*ast.MappingValueNode, 0, len(m.Items))
	for i, p := range m.Items {
		if p == nil {
			return nil, fmt.Errorf("index %d: pair must not be nil", i)
		}
		p.inFlow = m.node.IsFlowStyle
		mv, err := p.sync()
		if err != nil {
			return nil, fmt.Errorf("index %d: %w", i, err)
		}
		mv.IsFlowStyle = m.node.IsFlowStyle
		values = append(values, mv)
	}
	if m.comment != m.origComment {
		// The comment after the map is stored in the last pair.
		// Unless it's changed, it stays with the pair.
		m.origComment = m.comment
		if m.footPair != nil && m.footPair.node != nil {
			m.footPair.node.FootComment = nil
		}
		m.footPair = nil
		if len(values) > 0 {
			values[len(values)-1].FootComment = newCommentGroup(m.comment)
			m.footPair = m.Items[len(m.Items)-1]
		}
	}
	m.node.Values = values
	return m.outer, nil
}

func (p *Pair) sync() (*ast.MappingValueNode, error) {
	key, err := p.syncKey()
	if err != nil {
		return nil, err
	}
	if p.Value == nil {
		p.Value = &Scalar{}
	}
	vn, err := p.Value.sync(p)
	if err != nil {
		return nil, fmt.Errorf("key %s: %w", p.Key.String(), err)
	}
	if p.node == nil {
		p.node = ast.MappingValue(token.MappingValue(&token.Position{Column: key.GetToken().Position.Column}), key, vn)
	}
	p.node.Key = key
	p.node.Value = vn
	if p.commentBefore != p.origCommentBefore {
		p.origCommentBefore = p.commentBefore
		p.origCommentBeforeNode = newCommentGroup(p.commentBefore)
	}
	if err := p.node.SetComment(p.origCommentBeforeNode); err != nil {
		return nil, fmt.Errorf("set a comment: %w", err)
	}
	return p.node, nil
}

func (p *Pair) syncKey() (ast.MapKeyNode, error) {
	if p.Key == nil {
		return nil, errors.New("Pair.Key must not be nil")
	}
	var oldKeyToken *token.Token
	if p.node != nil && p.node.Key != nil {
		oldKeyToken = p.node.Key.GetToken()
	}
	kn, err := p.Key.sync(p)
	if err != nil {
		return nil, fmt.Errorf("key: %w", err)
	}
	key, ok := kn.(ast.MapKeyNode)
	if !ok {
		return nil, fmt.Errorf("unsupported map key type: %s", kn.Type())
	}
	if p.Key.standalone && oldKeyToken != nil {
		// Keep the position of the renamed key.
		tk := key.GetToken()
		pos := *oldKeyToken.Position
		tk.Position = &pos
		tk.Prev = oldKeyToken.Prev
		p.Key.standalone = false
	}
	p.Key.origParent = p
	return key, nil
}

func (s *Seq) sync(parent any) (ast.Node, error) {
	if err := s.initAST(); err != nil {
		return nil, err
	}
	if len(s.Items) > 0 {
		fixFlowStyle(&s.nodeBase, &s.node.IsFlowStyle, parent, len(s.node.Values) == 0)
	}
	n := len(s.Items)
	values := make([]ast.Node, n)
	heads := make([]*ast.CommentGroupNode, n)
	entries := make([]*ast.SequenceEntryNode, n)
	newEntries := make(map[Node]*ast.SequenceEntryNode, n)
	for i, item := range s.Items {
		if item == nil {
			item = &Scalar{}
			s.Items[i] = item
		}
		v, err := item.sync(s)
		if err != nil {
			return nil, fmt.Errorf("index %d: %w", i, err)
		}
		values[i] = v
		heads[i] = item.base().commentBeforeNode()
		entries[i] = s.entry(item, v, heads[i])
		newEntries[item] = entries[i]
	}
	s.entries = newEntries
	if !s.node.IsFlowStyle {
		// The comment before the first item is stored in the sequence itself.
		s.node.Comment = takeFirst(heads)
	}
	if s.comment != s.origComment {
		s.origComment = s.comment
		s.node.FootComment = newCommentGroup(s.comment)
	}
	s.node.Values = values
	s.node.ValueHeadComments = heads
	s.node.Entries = entries
	return s.outer, nil
}

// takeFirst returns the first comment and removes it from the list.
func takeFirst(heads []*ast.CommentGroupNode) *ast.CommentGroupNode {
	if len(heads) == 0 {
		return nil
	}
	h := heads[0]
	heads[0] = nil
	return h
}

// entry returns the sequence entry AST of the item.
func (s *Seq) entry(item Node, v ast.Node, head *ast.CommentGroupNode) *ast.SequenceEntryNode {
	entry, ok := s.entries[item]
	if !ok {
		entry = ast.SequenceEntry(token.SequenceEntry("-", &token.Position{}), v, nil)
	}
	entry.Value = v
	entry.HeadComment = head
	return entry
}

func (a *Alias) sync(_ any) (ast.Node, error) {
	if err := a.syncComment(a.node); err != nil {
		return nil, err
	}
	return a.outer, nil
}

type visitFunc func(n ast.Node)

func (f visitFunc) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	f(n)
	return f
}

// valueToAST converts a Go value to a YAML AST.
// The value is marshaled to YAML and parsed, so the positions of the tokens are consistent.
func valueToAST(v any) (ast.Node, error) {
	b, err := yaml.MarshalWithOptions(v, yaml.Indent(2), yaml.IndentSequence(true)) //nolint:mnd
	if err != nil {
		return nil, fmt.Errorf("marshal a value to YAML: %w", err)
	}
	return parseSnippet(b)
}

// parseSnippet parses a YAML text including a single document and returns the body.
func parseSnippet(b []byte) (ast.Node, error) {
	file, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}
	if len(file.Docs) != 1 {
		return nil, fmt.Errorf("YAML must have a single document, but got %d documents", len(file.Docs))
	}
	body := file.Docs[0].Body
	if body == nil {
		return ast.Null(token.New("null", "null", &token.Position{Column: 1})), nil
	}
	return body, nil
}
