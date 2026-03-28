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
