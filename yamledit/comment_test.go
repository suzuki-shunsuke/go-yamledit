package yamledit_test

import (
	"fmt"
	"log"

	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
)

func ExampleWithComment() {
	yml := `
- foo
`

	s, err := yamledit.EditBytes("example.yaml", []byte(yml), yamledit.ListAction(
		"$",
		// Add "zoo" with comment
		yamledit.AddValuesToList(1, yamledit.WithComment("zoo", " comment is added")),
	))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
	// Output:
	// - foo
	// - zoo # comment is added
}
