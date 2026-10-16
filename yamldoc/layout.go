package yamldoc

import (
	"strings"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

// goccy/go-yaml decides the indentation of the output by the column of tokens.
// layout fixes the columns of new and moved nodes.
// Nodes kept under the same parent keep their original columns, so the original indentation is kept as much as possible.

// style is the indentation style of a document.
type style struct {
	// indent is the indentation width of a nested map.
	indent int
	// seqIndent is the indentation width of a sequence under a map key.
	// It is 0 if sequences aren't indented.
	//   foo:
	//   - bar
	seqIndent int
}

const defaultIndent = 2

// detectStyle detects the indentation style from the original document.
func detectStyle(n ast.Node) *style {
	st := &style{indent: defaultIndent, seqIndent: defaultIndent}
	if n == nil {
		return st
	}
	foundMap, foundSeq := false, false
	ast.Walk(visitFunc(func(n ast.Node) {
		mv, ok := n.(*ast.MappingValueNode)
		if !ok || mv.IsFlowStyle || mv.Key == nil {
			return
		}
		if !foundMap {
			st.indent, foundMap = detectIndent(mv, st.indent)
		}
		if !foundSeq {
			st.seqIndent, foundSeq = detectSeqIndent(mv, st.seqIndent)
		}
	}), n)
	return st
}

// detectIndent returns the indentation width of the map nested in the pair.
// If it can't be detected, it returns def and false.
func detectIndent(mv *ast.MappingValueNode, def int) (int, bool) {
	var child *ast.MappingValueNode
	inner, _ := unwrapNode(mv.Value)
	switch v := inner.(type) {
	case *ast.MappingNode:
		if !v.IsFlowStyle && len(v.Values) > 0 {
			child = v.Values[0]
		}
	case *ast.MappingValueNode:
		if !v.IsFlowStyle {
			child = v
		}
	}
	if child == nil {
		return def, false
	}
	if d := child.Key.GetToken().Position.Column - mv.Key.GetToken().Position.Column; d > 0 {
		return d, true
	}
	return def, false
}

// detectSeqIndent returns the indentation width of the sequence nested in the pair.
// If it can't be detected, it returns def and false.
func detectSeqIndent(mv *ast.MappingValueNode, def int) (int, bool) {
	inner, _ := unwrapNode(mv.Value)
	v, ok := inner.(*ast.SequenceNode)
	if !ok || v.IsFlowStyle || len(v.Values) == 0 {
		return def, false
	}
	if d := v.Start.Position.Column - mv.Key.GetToken().Position.Column; d >= 0 {
		return d, true
	}
	return def, false
}

// layout fixes the columns of the node and its descendants.
// parent is the current parent of the node.
// target is the column where the node should be placed if it's new or moved.
// For a map, it's the column of keys.
// For a sequence, it's the column of "-".
// For a scalar, it's the column of the context (the key or the sequence item).
func (st *style) layout(n Node, parent any, target int) {
	if n == nil {
		return
	}
	b := n.base()
	kept := b.kept(parent)
	switch v := n.(type) {
	case *Scalar:
		if v.standalone {
			// The node was parsed from a standalone text, so it's at the column 1.
			shiftAST(v.outer, target-1)
			v.standalone = false
		}
	case *Map:
		st.layoutMap(v, kept, target)
	case *Seq:
		st.layoutSeq(v, kept, target)
	}
	b.origParent = parent
}

func keyColumn(p *Pair) int {
	return p.node.Key.GetToken().Position.Column
}

func (st *style) layoutMap(m *Map, kept bool, target int) {
	col := mapColumn(m, kept, target)
	for _, p := range m.Items {
		if !m.node.IsFlowStyle {
			shiftPair(p, col-keyColumn(p))
		}
		p.origParent = m
		p.Key.origParent = p
		p.Key.standalone = false
		st.layout(p.Value, p, st.valueColumn(p, col))
	}
}

// mapColumn returns the column of keys of the map.
// If the map is kept, the column of the existing keys is used.
func mapColumn(m *Map, kept bool, target int) int {
	if !kept {
		return target
	}
	for _, p := range m.Items {
		if p.origParent == m {
			return keyColumn(p)
		}
	}
	return target
}

// shiftPair shifts the key of the pair by d.
// The value is also shifted if it's placed relative to the key.
func shiftPair(p *Pair, d int) {
	if d == 0 {
		return
	}
	sh := newShifter(d)
	sh.node(p.node.Key)
	sh.token(p.node.Start)
	if vb := p.Value.base(); vb.kept(p) || (vb.origParent == nil && !vb.standalone) {
		// A value created with its parent is placed relative to the parent, so it's shifted together.
		sh.node(vb.outer)
	}
}

// valueColumn returns the column where the value of the pair should be placed.
// col is the column of the key.
func (st *style) valueColumn(p *Pair, col int) int {
	switch v := p.Value.(type) {
	case *Map:
		if len(v.Items) == 0 {
			inlineEmpty(p, v.node)
		}
		return col + st.indent
	case *Seq:
		if len(v.Items) == 0 {
			inlineEmpty(p, v.node)
		}
		return col + st.seqIndent
	default:
		return col
	}
}

// inlineEmpty makes an empty collection printed in the same line as the key.
// goccy/go-yaml prints the value in the next line if the indent level of the value is greater than the key's.
//
//	key: {}
func inlineEmpty(p *Pair, n ast.Node) {
	tk := n.GetToken()
	if tk == nil || tk.Position == nil {
		return
	}
	tk.Position.IndentLevel = min(tk.Position.IndentLevel, p.node.Key.GetToken().Position.IndentLevel)
}

func (st *style) layoutSeq(s *Seq, kept bool, target int) {
	col := target
	if s.node.Start != nil {
		col = s.node.Start.Position.Column
	}
	if !kept && !s.node.IsFlowStyle {
		shiftAST(s.outer, target-col)
		col = target
	}
	for _, item := range s.Items {
		st.layout(item, s, col+2) //nolint:mnd
	}
}

// shiftAST shifts the columns of all tokens in the node by d.
func shiftAST(n ast.Node, d int) {
	if d == 0 || n == nil {
		return
	}
	newShifter(d).node(n)
}

type shifter struct {
	d       int
	seen    map[*token.Position]struct{}
	origins map[*token.Token]struct{}
}

func newShifter(d int) *shifter {
	return &shifter{
		d:       d,
		seen:    map[*token.Position]struct{}{},
		origins: map[*token.Token]struct{}{},
	}
}

func (s *shifter) token(tk *token.Token) {
	if s.d == 0 || tk == nil || tk.Position == nil {
		return
	}
	if _, ok := s.seen[tk.Position]; ok {
		return
	}
	s.seen[tk.Position] = struct{}{}
	tk.Position.Column += s.d
}

func (s *shifter) node(n ast.Node) {
	if s.d == 0 || n == nil {
		return
	}
	ast.Walk(visitFunc(func(n ast.Node) {
		s.token(n.GetToken())
		for _, tk := range extraTokens(n) {
			s.token(tk)
		}
		if v, ok := n.(*ast.LiteralNode); ok && v.Value != nil {
			s.origin(v.Value.GetToken())
		}
	}), n)
}

// extraTokens returns tokens of the node other than GetToken.
func extraTokens(n ast.Node) []*token.Token {
	switch v := n.(type) {
	case *ast.MappingNode:
		return []*token.Token{v.Start, v.End}
	case *ast.MappingValueNode:
		return []*token.Token{v.Start}
	case *ast.SequenceNode:
		tokens := []*token.Token{v.Start, v.End}
		for _, e := range v.Entries {
			tokens = append(tokens, e.Start)
		}
		return tokens
	case *ast.AnchorNode:
		return []*token.Token{v.Start}
	case *ast.TagNode:
		return []*token.Token{v.Start}
	case *ast.AliasNode:
		return []*token.Token{v.Start}
	case *ast.LiteralNode:
		return []*token.Token{v.Start}
	default:
		return nil
	}
}

// origin shifts the indentation of a block scalar.
// The output of a block scalar is its original text, so the text itself must be changed.
func (s *shifter) origin(tk *token.Token) {
	if tk == nil {
		return
	}
	if _, ok := s.origins[tk]; ok {
		return
	}
	s.origins[tk] = struct{}{}
	s.token(tk)
	lines := strings.Split(tk.Origin, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if s.d > 0 {
			lines[i] = strings.Repeat(" ", s.d) + line
			continue
		}
		spaces := len(line) - len(strings.TrimLeft(line, " "))
		lines[i] = line[min(spaces, -s.d):]
	}
	tk.Origin = strings.Join(lines, "\n")
}
