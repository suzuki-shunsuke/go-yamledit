package yamldoc

import (
	"fmt"

	"github.com/goccy/go-yaml/ast"
)

// Node is a node of a YAML document.
// The concrete type is one of *Scalar, *Map, *Seq, and *Alias.
// Use a type switch or a type assertion to access the concrete type, like isScalar, isMap, and isSeq of eemeli/yaml.
type Node interface {
	// Comment returns the comment after the node.
	// For *Scalar and *Alias, this is the line comment.
	// For *Map and *Seq, this is the comment after the last item.
	Comment() string
	// SetComment sets the comment after the node.
	// An empty string removes the comment.
	SetComment(comment string)
	// CommentBefore returns the comment before the node.
	// This is used only when the node is an item of a sequence.
	// Use Pair.CommentBefore for items of a map.
	CommentBefore() string
	// SetCommentBefore sets the comment before the node.
	// An empty string removes the comment.
	SetCommentBefore(comment string)
	// Anchor returns the name of the anchor of the node.
	// If the node has no anchor, it returns an empty string.
	Anchor() string
	// Decode decodes the node to a Go value like yaml.Unmarshal.
	Decode(v any) error

	base() *nodeBase
	// sync writes the changes of the node back to the AST and returns the AST node.
	// parent is the current parent of the node.
	sync(parent any) (ast.Node, error)
}

// nodeBase is embedded in all nodes.
type nodeBase struct {
	comment       string
	commentBefore string

	// The original values, which are used to detect changes.
	// The AST is updated only if the value is changed so that the original format is kept as much as possible.
	origComment           string
	origCommentBefore     string
	origCommentBeforeNode *ast.CommentGroupNode

	// origParent is the parent (*Document, *Pair, or *Seq) when the node was laid out last time.
	// It is nil if the node was created newly.
	// If the node is still under the same parent, its indentation is kept.
	origParent any
	// standalone is true if the AST node was generated from a standalone YAML text.
	// Then the AST node is at the column 1, so it must be shifted to the right position.
	standalone bool

	// outer is the outermost AST node, which may be an anchor or tag node wrapping the actual node.
	outer ast.Node
	// holder is the anchor or tag node directly wrapping the actual node.
	holder ast.Node
}

func (b *nodeBase) Comment() string                 { return b.comment }
func (b *nodeBase) SetComment(comment string)       { b.comment = comment }
func (b *nodeBase) CommentBefore() string           { return b.commentBefore }
func (b *nodeBase) SetCommentBefore(comment string) { b.commentBefore = comment }

func (b *nodeBase) Anchor() string {
	n := b.outer
	for {
		switch v := n.(type) {
		case *ast.AnchorNode:
			return v.Name.String()
		case *ast.TagNode:
			n = v.Value
		default:
			return ""
		}
	}
}

func (b *nodeBase) base() *nodeBase { return b }

// setInner replaces the actual AST node.
func (b *nodeBase) setInner(n ast.Node) {
	switch h := b.holder.(type) {
	case *ast.AnchorNode:
		h.Value = n
	case *ast.TagNode:
		h.Value = n
	default:
		b.outer = n
	}
}

// commentBeforeNode returns the comment before the node as AST.
// The original AST is reused if the comment isn't changed.
func (b *nodeBase) commentBeforeNode() *ast.CommentGroupNode {
	if b.commentBefore != b.origCommentBefore {
		b.origCommentBefore = b.commentBefore
		b.origCommentBeforeNode = newCommentGroup(b.commentBefore)
	}
	return b.origCommentBeforeNode
}

func (b *nodeBase) kept(parent any) bool {
	return b.origParent != nil && b.origParent == parent
}

// Scalar is a scalar node such as a string, a number, a boolean, and null.
type Scalar struct {
	nodeBase

	// Value is the value of the scalar.
	// It is a string, int, float64, bool, or nil.
	// Integers are normalized to int if the value is in the range of int.
	// You can set any Go value that is marshaled to a YAML scalar.
	Value any

	orig any
	node ast.Node
}

// String returns the value as a string.
// If the value is nil, it returns an empty string.
func (s *Scalar) String() string {
	if s.Value == nil {
		return ""
	}
	return fmt.Sprint(s.Value)
}

// Map is a mapping node.
type Map struct {
	nodeBase

	// Items is the list of key-value pairs.
	// You can modify this slice directly.
	// e.g. slices.Insert, slices.Delete, slices.SortFunc
	Items []*Pair

	node *ast.MappingNode
	// footPair is the pair having the comment after the map.
	footPair *Pair
}

// Pair is a key-value pair of a map.
type Pair struct {
	// Key is the key.
	// You can rename the key by changing Key.Value.
	Key *Scalar
	// Value is the value.
	// If it is nil, it is treated as null.
	Value Node

	commentBefore         string
	origCommentBefore     string
	origCommentBeforeNode *ast.CommentGroupNode

	origParent *Map
	node       *ast.MappingValueNode
	// inFlow is true if the pair is in a flow map.
	inFlow bool
}

// CommentBefore returns the comment before the pair.
func (p *Pair) CommentBefore() string { return p.commentBefore }

// SetCommentBefore sets the comment before the pair.
// An empty string removes the comment.
func (p *Pair) SetCommentBefore(comment string) { p.commentBefore = comment }

// KeyString returns the key as a string.
func (p *Pair) KeyString() string {
	if p.Key == nil {
		return ""
	}
	return p.Key.String()
}

// Seq is a sequence node.
type Seq struct {
	nodeBase

	// Items is the list of items.
	// You can modify this slice directly.
	// e.g. slices.Insert, slices.Delete, slices.SortFunc
	Items []Node

	node    *ast.SequenceNode
	entries map[Node]*ast.SequenceEntryNode
}

// Alias is an alias node such as `*foo`.
// Alias is read only.
type Alias struct {
	nodeBase

	node *ast.AliasNode
}

// Name returns the name of the anchor the alias refers to.
func (a *Alias) Name() string {
	return a.node.Value.String()
}
