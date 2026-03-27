package mag_test

import (
	"fmt"
	"log"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleAddValuesToList() {
	yml := `
- foo # comment
- bar
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	act := mag.List("$", mag.AddValuesToList(0, "zoo"))
	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// - zoo
	// - foo # comment
	// - bar
}

func ExampleAddValuesToList_negative_index() {
	yml := `
- foo # comment
- bar
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	// Add "zoo" to the last position
	act := mag.List("$", mag.AddValuesToList(-1, "zoo"))
	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// - foo # comment
	// - bar
	// - zoo
}
