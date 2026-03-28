package yamledit_test

import (
	"fmt"
	"log"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
)

func ExampleWithComment() {
	yml := `
- foo
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	act := yamledit.ListAction(
		"$",
		// Add "zoo" with comment
		yamledit.AddValuesToList(1, yamledit.WithComment("zoo", " comment is added")),
	)
	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// - foo
	// - zoo # comment is added
}
