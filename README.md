# go-yamledit

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/suzuki-shunsuke/go-yamledit)
[![Go Reference](https://pkg.go.dev/badge/github.com/suzuki-shunsuke/go-yamledit.svg)](https://pkg.go.dev/github.com/suzuki-shunsuke/go-yamledit/yamledit)

go-yamledit is a Go library to edit YAML files while keeping YAML comments and indentation.
It provides high-level API to edit YAML using [goccy/go-yaml](https://github.com/goccy/go-yaml).
If you need more flexibility and performance, use goccy/go-yaml directly.
This package allows you to edit YAML files easily without operating complicated YAML AST.

goccy/go-yaml is an excellent library for editing YAML using an AST.
However, using it effectively is not always straightforward.
While it provides a large number of APIs, the documentation and example code are not particularly comprehensive.
In practice, you often need to inspect the parsed `ast.Node` structure and figure out how to implement the desired changes through trial and debugging.
Even for small migrations, this can make the task feel unnecessarily heavy.

To address these challenges, go-yamledit provides high-level APIs that make common use cases easy to implement.
At the same time, it remains flexible enough to support a wide range of scenarios.

[For full document, please see GoDoc.](https://pkg.go.dev/github.com/suzuki-shunsuke/go-yamledit/yamledit)

## yamldoc: Document API like eemeli/yaml

[![Go Reference](https://pkg.go.dev/badge/github.com/suzuki-shunsuke/go-yamledit.svg)](https://pkg.go.dev/github.com/suzuki-shunsuke/go-yamledit/yamldoc)

The package [yamldoc](yamldoc) provides another API like the Document API of [eemeli/yaml](https://github.com/eemeli/yaml).
While yamledit applies declarative actions to nodes selected by YAML paths, yamldoc converts the AST to a simple tree (`Map`, `Seq`, `Pair`, `Scalar`), and you edit the tree with ordinary Go code.
Structural changes can be done by modifying `Map.Items` and `Seq.Items` with the standard `slices` package.
The changes are written back to the AST and the indentation is fixed when you call `ToString`.

- Use yamledit to apply the same change to many places selected by YAML paths.
- Use yamldoc to edit a document procedurally, e.g. inserting, reordering, and renaming keys.

### Example

```go
doc, err := yamldoc.Parse([]byte(`# user
name: foo # name
id: 1
tags:
  - b
  - a
`))
if err != nil {
	return err
}
// Update a value keeping the comment.
if err := doc.Set("name", "bar"); err != nil {
	return err
}
// Delete a key.
doc.Delete("id")
// Add a key.
if err := doc.Set("age", 20); err != nil {
	return err
}

m := doc.Contents.(*yamldoc.Map)
// Insert a pair at a specific position.
p, err := yamldoc.NewPair("id", 123)
if err != nil {
	return err
}
m.Items = slices.Insert(m.Items, m.IndexOf("name")+1, p)
// Rename a key.
m.Pair("age").Key.Value = "years"

// Sort a sequence and add a comment.
tags := doc.Get("tags").(*yamldoc.Seq)
slices.SortFunc(tags.Items, func(a, b yamldoc.Node) int {
	return cmp.Compare(a.(*yamldoc.Scalar).String(), b.(*yamldoc.Scalar).String())
})
tags.Items[1].SetComment(" tag b")

s, err := doc.ToString()
if err != nil {
	return err
}
fmt.Print(s)
```

Output:

```yaml
# user
name: bar # name
id: 123
tags:
  - a
  - b # tag b
years: 20
```

### Comparison with eemeli/yaml

| eemeli/yaml | go-yamledit |
|---|---|
| `YAML.parseDocument(src)` | `yamldoc.Parse(b)` |
| `YAML.parseAllDocuments(src)` | `yamldoc.ParseAll(b)` |
| `doc.toString()` | `doc.ToString()`, `file.ToString()` |
| `doc.contents` | `doc.Contents` |
| `doc.get` / `set` / `has` / `delete` | `Get` / `Set` / `Has` / `Delete` |
| `doc.getIn` / `setIn` / `hasIn` / `deleteIn` | `GetIn` / `SetIn` / `HasIn` / `DeleteIn` |
| `doc.createNode(v)` | `yamldoc.NewNode(v)` |
| `doc.createPair(k, v)` | `yamldoc.NewPair(k, v)` |
| `isMap(n)`, `isSeq(n)`, `isScalar(n)` | type switch or type assertion: `n.(*yamldoc.Map)` |
| `map.items.splice(...)` | `slices.Insert`, `slices.Delete` on `Map.Items` |
| `pair.key.value = "new"` | `pair.Key.Value = "new"` |
| `node.comment`, `node.commentBefore` | `Comment()` / `SetComment()`, `CommentBefore()` / `SetCommentBefore()` |
| `node.toJS()` | `node.Decode(&v)`, `yamldoc.GetAs[T](node, path...)` |

`EditFile` reads a file, edits it, and writes it back:

```go
changed, err := yamldoc.EditFile("workflow.yaml", func(f *yamldoc.File) error {
	for _, doc := range f.Docs {
		if err := doc.SetIn([]any{"permissions", "contents"}, "read"); err != nil {
			return err
		}
	}
	return nil
})
```

### Indentation

New and moved nodes are indented according to the style detected from the document: the indentation width of nested maps, and whether sequences under keys are indented.
Nodes staying under the same parent keep their original columns.

### Limitations

- Aliases are read only. Anchors can be read by `Anchor()`.
- Only scalar map keys are supported. Complex keys (`? key`) cause a parse error.
- Go maps are marshaled in the order of keys. Use `yaml.MapSlice` to keep the order.
- Values are marshaled by goccy/go-yaml, so strings like `y`, `n`, `on`, and `off` are quoted for YAML 1.1 compatibility.
- Some YAML isn't handled well by goccy/go-yaml itself, for example a comment in a multi-line flow sequence, or a tag without a value such as `!!python/name:foo` followed by sequence items.
