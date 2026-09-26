package yamldoc

import (
	"fmt"

	"github.com/goccy/go-yaml/ast"
)

// unwrapNode removes anchor and tag nodes wrapping the actual node.
// It returns the actual node and the anchor or tag node directly wrapping it.
func unwrapNode(n ast.Node) (ast.Node, ast.Node) {
	var holder ast.Node
	for {
		switch v := n.(type) {
		case *ast.AnchorNode:
			holder = v
			n = v.Value
		case *ast.TagNode:
			holder = v
			n = v.Value
		default:
			return n, holder
		}
	}
}

// wrap converts an AST node to a Node.
// parent is the parent of the node.
// If fresh is true, the node and its descendants are treated as newly created nodes,
// which are laid out according to the document's indentation style.
func wrap(n ast.Node, parent any, fresh bool) (Node, error) {
	inner, holder := unwrapNode(n)
	b := nodeBase{
		outer:  n,
		holder: holder,
	}
	if !fresh {
		b.origParent = parent
	}
	switch v := inner.(type) {
	case *ast.MappingNode:
		return wrapMap(v, b, fresh)
	case *ast.MappingValueNode:
		// A map with a single pair may be parsed as MappingValueNode.
		m := ast.Mapping(v.GetToken(), false, v)
		b.setInner(m)
		return wrapMap(m, b, fresh)
	case *ast.SequenceNode:
		return wrapSeq(v, b, fresh)
	case *ast.AliasNode:
		b.comment = commentString(v.GetComment())
		b.origComment = b.comment
		return &Alias{nodeBase: b, node: v}, nil
	case ast.ScalarNode:
		return wrapScalar(v, b), nil
	case nil:
		return nil, nil //nolint:nilnil
	default:
		return nil, fmt.Errorf("unsupported node type: %s", inner.Type())
	}
}

func wrapScalar(n ast.ScalarNode, b nodeBase) *Scalar {
	b.comment = commentString(n.GetComment())
	b.origComment = b.comment
	v := scalarValue(n)
	return &Scalar{
		nodeBase: b,
		Value:    v,
		orig:     v,
		node:     n,
	}
}

func scalarValue(n ast.ScalarNode) any {
	if l, ok := n.(*ast.LiteralNode); ok {
		// LiteralNode.GetValue returns the YAML text including the header "|".
		if l.Value != nil {
			return l.Value.Value
		}
		return ""
	}
	return normalizeValue(n.GetValue())
}

func wrapMap(n *ast.MappingNode, b nodeBase, fresh bool) (*Map, error) {
	m := &Map{
		nodeBase: b,
		Items:    make([]*Pair, 0, len(n.Values)),
		node:     n,
	}
	if len(n.Values) > 0 {
		// The comment after the map is stored in the last pair.
		m.comment = commentString(n.Values[len(n.Values)-1].FootComment)
		m.origComment = m.comment
	}
	for _, mv := range n.Values {
		p, err := wrapPair(mv, m, fresh)
		if err != nil {
			return nil, err
		}
		m.Items = append(m.Items, p)
	}
	if len(m.Items) > 0 {
		m.footPair = m.Items[len(m.Items)-1]
	}
	return m, nil
}

func wrapPair(mv *ast.MappingValueNode, m *Map, fresh bool) (*Pair, error) {
	p := &Pair{
		node:                  mv,
		commentBefore:         commentString(mv.GetComment()),
		origCommentBeforeNode: mv.GetComment(),
	}
	p.origCommentBefore = p.commentBefore
	if !fresh {
		p.origParent = m
	}
	kn, holder := unwrapNode(mv.Key)
	sk, ok := kn.(ast.ScalarNode)
	if !ok {
		return nil, fmt.Errorf("unsupported map key type: %s", kn.Type())
	}
	kb := nodeBase{outer: mv.Key, holder: holder}
	if !fresh {
		kb.origParent = p
	}
	p.Key = wrapScalar(sk, kb)
	v, err := wrap(mv.Value, p, fresh)
	if err != nil {
		return nil, fmt.Errorf("key %s: %w", p.Key.String(), err)
	}
	p.Value = v
	return p, nil
}

func wrapSeq(n *ast.SequenceNode, b nodeBase, fresh bool) (*Seq, error) {
	s := &Seq{
		nodeBase: b,
		Items:    make([]Node, 0, len(n.Values)),
		node:     n,
		entries:  make(map[Node]*ast.SequenceEntryNode, len(n.Values)),
	}
	s.comment = commentString(n.FootComment)
	s.origComment = s.comment
	for i, value := range n.Values {
		item, err := wrap(value, s, fresh)
		if err != nil {
			return nil, fmt.Errorf("index %d: %w", i, err)
		}
		ib := item.base()
		var cg *ast.CommentGroupNode
		if len(n.ValueHeadComments) == len(n.Values) {
			cg = n.ValueHeadComments[i]
		}
		if i == 0 && !n.IsFlowStyle && n.Comment != nil {
			// The comment before the first item is stored in the sequence itself.
			cg = n.Comment
		}
		ib.commentBefore = commentString(cg)
		ib.origCommentBefore = ib.commentBefore
		ib.origCommentBeforeNode = cg
		if i < len(n.Entries) {
			s.entries[item] = n.Entries[i]
		}
		s.Items = append(s.Items, item)
	}
	return s, nil
}
