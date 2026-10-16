package yamldoc

import (
	"strings"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/token"
)

// commentString converts a comment AST to a string.
// Each line doesn't include "#".
// e.g. "# foo\n# bar" => " foo\n bar"
func commentString(cg *ast.CommentGroupNode) string {
	if cg == nil {
		return ""
	}
	lines := make([]string, 0, len(cg.Comments))
	for _, c := range cg.Comments {
		if c == nil || c.Token == nil {
			continue
		}
		lines = append(lines, c.Token.Value)
	}
	return strings.Join(lines, "\n")
}

// newCommentGroup converts a string to a comment AST.
// "#" is prepended to each line.
// e.g. " foo\n bar" => "# foo\n# bar"
func newCommentGroup(s string) *ast.CommentGroupNode {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	tokens := make([]*token.Token, len(lines))
	for i, line := range lines {
		tokens[i] = token.Comment(line, "#"+line, &token.Position{})
	}
	return ast.CommentGroup(tokens)
}
