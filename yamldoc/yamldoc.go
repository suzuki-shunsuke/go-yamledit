// Package yamldoc provides an API to edit YAML documents like eemeli/yaml's Document API, keeping comments and indentation.
// It is a wrapper of goccy/go-yaml.
//
// Parse a YAML text to a tree of Node, edit the tree, and serialize it by ToString.
// Structural changes can be done by modifying Map.Items and Seq.Items directly with the standard slices package.
// The changes are written back to the goccy/go-yaml's AST when ToString is called.
package yamldoc

import (
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

// File is a YAML file, which may have multiple documents.
type File struct {
	Docs []*Document

	file *ast.File
}

// Document is a YAML document.
type Document struct {
	// Contents is the root node of the document.
	// It is nil if the document is empty.
	Contents Node

	node  *ast.DocumentNode
	style *style
	// comment is the body of a document having only comments.
	comment *ast.CommentGroupNode
}

// ParseAll parses a YAML text including multiple documents.
func ParseAll(b []byte) (*File, error) {
	file, err := parser.ParseBytes(b, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}
	f := &File{
		Docs: make([]*Document, len(file.Docs)),
		file: file,
	}
	for i, d := range file.Docs {
		doc, err := newDocument(d)
		if err != nil {
			return nil, fmt.Errorf("document %d: %w", i, err)
		}
		f.Docs[i] = doc
	}
	return f, nil
}

// Parse parses a YAML text including a single document.
// If the text includes multiple documents, it returns an error.
// Use ParseAll for multiple documents.
func Parse(b []byte) (*Document, error) {
	f, err := ParseAll(b)
	if err != nil {
		return nil, err
	}
	switch len(f.Docs) {
	case 0:
		return newDocument(ast.Document(nil, nil))
	case 1:
		return f.Docs[0], nil
	default:
		return nil, fmt.Errorf("YAML must have a single document, but got %d documents. Use ParseAll", len(f.Docs))
	}
}

func newDocument(d *ast.DocumentNode) (*Document, error) {
	doc := &Document{
		node:  d,
		style: detectStyle(d.Body),
	}
	if cg, ok := d.Body.(*ast.CommentGroupNode); ok {
		// The document has only comments.
		doc.comment = cg
		return doc, nil
	}
	n, err := wrap(d.Body, doc, false)
	if err != nil {
		return nil, err
	}
	doc.Contents = n
	return doc, nil
}

// ToString returns the YAML text of the file.
func (f *File) ToString() (string, error) {
	for i, doc := range f.Docs {
		if err := doc.sync(); err != nil {
			return "", fmt.Errorf("document %d: %w", i, err)
		}
	}
	docs := make([]*ast.DocumentNode, len(f.Docs))
	for i, doc := range f.Docs {
		docs[i] = doc.node
	}
	f.file.Docs = docs
	return f.file.String(), nil
}

// ToString returns the YAML text of the document.
func (d *Document) ToString() (string, error) {
	if err := d.sync(); err != nil {
		return "", err
	}
	s := d.node.String()
	if s == "" {
		return "", nil
	}
	return s + "\n", nil
}

func (d *Document) sync() error {
	if d.Contents == nil {
		if d.comment != nil {
			d.node.Body = d.comment
		} else {
			d.node.Body = nil
		}
		return nil
	}
	if d.comment != nil {
		// Contents are added to a document having only comments.
		// Keep the comments before the contents.
		d.prependComment()
	}
	n, err := d.Contents.sync(d)
	if err != nil {
		return err
	}
	d.node.Body = n
	d.style.layout(d.Contents, d, 1)
	fixClipLiterals(d.Contents, lastScalar(d.Contents))
	return nil
}

// lastScalar returns the last scalar in the document order.
func lastScalar(n Node) *Scalar {
	switch v := n.(type) {
	case *Scalar:
		return v
	case *Map:
		if len(v.Items) > 0 {
			return lastScalar(v.Items[len(v.Items)-1].Value)
		}
	case *Seq:
		if len(v.Items) > 0 {
			return lastScalar(v.Items[len(v.Items)-1])
		}
	}
	return nil
}

// fixClipLiterals adds the strip chomping indicator "-" to block scalars without a trailing line break.
// A block scalar at the end of a file without a line break has no trailing line break even with the clip chomping "|".
// If it's moved and followed by other nodes, a line break would be added to the value unexpectedly.
func fixClipLiterals(n Node, last *Scalar) {
	switch v := n.(type) {
	case *Scalar:
		if v != last {
			stripClip(v)
		}
	case *Map:
		for _, p := range v.Items {
			fixClipLiterals(p.Value, last)
		}
	case *Seq:
		for _, item := range v.Items {
			fixClipLiterals(item, last)
		}
	}
}

// EditFile reads a YAML file, edits it by the given function, and writes it back.
// It returns true if the file is changed.
func EditFile(filePath string, edit func(f *File) error) (bool, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("read a file: %w", err)
	}
	f, err := ParseAll(b)
	if err != nil {
		return false, err
	}
	if err := edit(f); err != nil {
		return false, err
	}
	s, err := f.ToString()
	if err != nil {
		return false, err
	}
	if s == string(b) {
		return false, nil
	}
	fi, err := os.Stat(filePath)
	if err != nil {
		return false, fmt.Errorf("get a file stat: %w", err)
	}
	if err := os.WriteFile(filePath, []byte(s), fi.Mode()); err != nil { //nolint:gosec
		return false, fmt.Errorf("write a file: %w", err)
	}
	return true, nil
}

func (d *Document) prependComment() {
	c := commentString(d.comment)
	switch v := d.Contents.(type) {
	case *Map:
		if len(v.Items) > 0 && v.Items[0] != nil {
			v.Items[0].SetCommentBefore(joinComments(c, v.Items[0].CommentBefore()))
			d.comment = nil
		}
	case *Seq:
		if len(v.Items) > 0 && v.Items[0] != nil {
			v.Items[0].SetCommentBefore(joinComments(c, v.Items[0].CommentBefore()))
			d.comment = nil
		}
	}
}

func joinComments(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	return a + "\n" + b
}

func stripClip(s *Scalar) {
	l, ok := s.node.(*ast.LiteralNode)
	if !ok || l.Start == nil || l.Value == nil || strings.HasSuffix(l.Value.Value, "\n") {
		return
	}
	if h := l.Start.Value; h != "" && !strings.ContainsAny(h, "-+") {
		l.Start.Value = h[:1] + "-" + h[1:]
	}
}
