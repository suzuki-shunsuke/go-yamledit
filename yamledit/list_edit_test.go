package yamledit_test

import (
	"fmt"
	"log"

	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
)

func ExampleEditListAction() {
	yml := `
- foo
- bar # comment
`

	s, err := yamledit.EditBytes("example.yaml", []byte(yml), yamledit.ListAction(
		"$",
		yamledit.EditListAction[string](
			func(m *yamledit.List[string]) error {
				return yamledit.RemoveValuesFromSequenceNode(m.Node, 0)
			},
		),
	))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
	// Output:
	// - bar # comment
}
