package yamledit_test

import (
	"fmt"
	"log"

	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
)

func ExampleRemoveValuesFromList() {
	yml := `
children:
  - foo # comment
  - bar # comment 2
`

	s, err := yamledit.EditBytes("example.yaml", []byte(yml), yamledit.ListAction(
		"$.children",
		// Remove foo
		yamledit.RemoveValuesFromList[string](func(value *yamledit.Node[string]) (bool, error) {
			return value.Value == "foo", nil
		}),
	))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
	// Output:
	// children:
	//   - bar # comment 2
}
