package mag_test

import (
	"fmt"
	"log"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/go-yamledit/mag"
)

func ExampleRemoveValuesFromList() {
	yml := `
children:
  - foo # comment
  - bar # comment 2
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	act := mag.ListAction(
		"$.children",
		// Remove foo
		mag.RemoveValuesFromList[string](func(value *mag.Node[string]) (bool, error) {
			return value.Value == "foo", nil
		}),
	)
	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// children:
	//   - bar # comment 2
}
