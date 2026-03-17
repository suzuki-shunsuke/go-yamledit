package mag

import (
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

type valueWithComment struct {
	Value   any
	Comment string
}

func WithComment(v any, comment string) any {
	return &valueWithComment{
		Value: v, Comment: comment,
	}
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
