package yamledit_test

import (
	"fmt"
	"log"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
)

func ExampleNewBytes() {
	yml := `
- foo # comment
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	act := yamledit.ListAction("$", yamledit.AddValuesToList(0, yamledit.NewBytes([]byte("hello # world"))))
	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// - hello # world
	// - foo # comment
}
