package yamledit

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

// EditFile is a helper function that reads a YAML file, applies actions to its AST, and writes it back.
func EditFile(path string, actions ...Action) error {
	f, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}
	file, err := parser.ParseFile(path, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse file: %w", err)
	}
	for _, doc := range file.Docs {
		for _, act := range actions {
			if err := act.Run(doc.Body); err != nil {
				return fmt.Errorf("run action: %w", err)
			}
		}
	}
	if err := os.WriteFile(path, []byte(file.String()), f.Mode()); err != nil { //nolint:gosec,mnd
		return fmt.Errorf("edit file: %w", err)
	}
	return nil
}

// EditBytes is a helper function that parses a YAML, applies actions to its AST, and returns the modified YAML string.
func EditBytes(b []byte, actions ...Action) (string, error) {
	file, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("parse YAML: %w", err)
	}
	for _, doc := range file.Docs {
		for _, act := range actions {
			if err := act.Run(doc.Body); err != nil {
				return "", fmt.Errorf("run action: %w", err)
			}
		}
	}
	return file.String(), nil
}

// BytesToNode parses the given YAML bytes and returns the root node.
// Returns an error if the bytes cannot be parsed.
// YAML should be a single document.
func BytesToNode(b []byte) (ast.Node, error) {
	file, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return file.Docs[0].Body, nil
}

// Bytes holds a YAML document.
// This is converted to ast.Node internally.
type Bytes struct {
	b      []byte
	isList bool
}

// NewBytes converts yaml bytes to a YAMLBytes struct.
// This is useful to pass YAML string to functions without temporary variables and error handling.
func NewBytes(b []byte) *Bytes {
	return &Bytes{b: b}
}

// NewListBytes converts a YAML list bytes to a Bytes struct.
// This is useful to pass YAML list string to functions without temporary variables and error handling.
func NewListBytes(b []byte) *Bytes {
	return &Bytes{
		b:      b,
		isList: true,
	}
}

type noop struct{}

// NoChange is a sentinel value that indicates no change should be made against mapping key or value.
var NoChange = noop{} //nolint:gochecknoglobals

func flatten(node ast.Node, depth int) ([]ast.Node, error) {
	if depth == 0 {
		return []ast.Node{node}, nil
	}

	seq, ok := node.(*ast.SequenceNode)
	if !ok {
		if depth == -1 {
			return []ast.Node{node}, nil
		}
		return nil, fmt.Errorf("expected a sequence node: %s", node.Type().String())
	}
	ret := []ast.Node{}
	for _, elem := range seq.Values {
		n := depth - 1
		if depth == -1 {
			n = -1
		}
		nodes, err := flatten(elem, n)
		if err != nil {
			return nil, err
		}
		ret = append(ret, nodes...)
	}
	return ret, nil
}

func getDepthByPath(yamlPath string) int { //nolint:cyclop
	count := 0
	inQuote := false
	for i := 0; i < len(yamlPath); i++ {
		ch := yamlPath[i]
		if ch == '\\' && inQuote && i+1 < len(yamlPath) && yamlPath[i+1] == '\'' {
			i++ // skip escaped quote
			continue
		}
		if ch == '\'' {
			inQuote = !inQuote
			continue
		}
		if inQuote {
			continue
		}
		if i+3 <= len(yamlPath) && yamlPath[i:i+3] == "[*]" {
			count++
			i += 2 // skip rest of [*]
			continue
		}
		if ch == '.' && i+1 < len(yamlPath) && yamlPath[i+1] == '.' {
			count++
			i++ // skip second dot
			continue
		}
	}
	return count
}

func valueToNode(value any) (ast.Node, error) {
	if node, ok := value.(ast.Node); ok {
		return node, nil
	}
	if b, ok := value.(*Bytes); ok {
		return BytesToNode(b.b)
	}
	valWithComment := toValueWithComment(value)
	v, err := yaml.ValueToNode(valWithComment.Value)
	if err != nil {
		return nil, err
	}
	if valWithComment.Comment == "" {
		return v, nil
	}
	if err := v.SetComment(commentGroupFromString(valWithComment.Comment)); err != nil {
		return nil, err
	}
	return v, nil
}
