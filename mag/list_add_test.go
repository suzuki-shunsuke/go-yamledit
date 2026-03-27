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
	act := mag.ListAction("$", mag.AddValuesToList(0, "zoo"))
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
	act := mag.ListAction("$", mag.AddValuesToList(-1, "zoo"))
	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// - foo # comment
	// - bar
	// - zoo
}

func ExampleAddValuesToList_with_list_bytes() {
	yml := `
- foo # comment
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	act := mag.ListAction("$", mag.AddValuesToList(-1, mag.NewListBytes([]byte(`
- bar # comment 1
- zoo # comment 2
`))))
	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// - foo # comment
	// - bar # comment 1
	// - zoo # comment 2
}
