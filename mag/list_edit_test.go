package mag_test

import (
	"fmt"
	"log"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleNewEditList() {
	yml := `
- foo
- bar # comment
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	act := mag.ListAction(
		"$",
		mag.NewEditList[string](
			func(m *mag.List[string]) error {
				return mag.RemoveValuesFromSequenceNode(m.Node, 0)
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
