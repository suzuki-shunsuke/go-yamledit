package yamledit_test

import (
	"fmt"
	"log"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
)

func ExampleEditListAction() {
	yml := `
- foo
- bar # comment
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	act := yamledit.ListAction(
		"$",
		yamledit.EditListAction[string](
			func(m *yamledit.List[string]) error {
				return yamledit.RemoveValuesFromSequenceNode(m.Node, 0)
			},
		),
	)
	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// - bar # comment
}
