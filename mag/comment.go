package mag

import (
	"fmt"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

// WithComment adds a comment to a value.
func WithComment(v any, comment string) any {
	if node, ok := v.(ast.Node); ok {
		if comment == "" {
			// remove comment from node
			node.SetComment(nil) //nolint:errcheck
			return node
		}
		node.SetComment(commentGroupFromString(comment)) //nolint:errcheck
		return node
	}
	if a, ok := v.(*valueWithComment); ok {
		return &valueWithComment{
			Value:   a.Value,
			Comment: comment,
		}
	}
	return &valueWithComment{
		Value: v, Comment: comment,
	}
}

type valueWithComment struct {
	Value   any
	Comment string
}

func commentGroupFromString(s string) *ast.CommentGroupNode {
	tk := token.Comment(s, "# "+s, &token.Position{})
	return ast.CommentGroup([]*token.Token{tk})
}

func toValueWithComment(v any) *valueWithComment {
	a, ok := v.(*valueWithComment)
	if ok {
		return a
	}
	return &valueWithComment{
		Value: v,
	}
}

func getComment(node ast.Node) string {
	if node == nil {
		return ""
	}
	cn := node.GetComment()
	if cn == nil {
		return ""
	}
	return cn.String()
}

type ChangeRemoveComment struct {
	Node ast.Node
}

func (a *ChangeRemoveComment) Run() error {
	if err := a.Node.SetComment(nil); err != nil {
		return fmt.Errorf("remove comment: %w", err)
	}
	return nil
}

type ChangeSetComment struct {
	Node    ast.Node
	Comment string
}

func (a *ChangeSetComment) Run() error {
	if err := a.Node.SetComment(commentGroupFromString(a.Comment)); err != nil {
		return fmt.Errorf("set comment: %w", err)
	}
	return nil
}
