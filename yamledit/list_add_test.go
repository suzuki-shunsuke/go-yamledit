package yamledit_test

import (
	"fmt"
	"log"

	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
)

func ExampleAddValuesToList() {
	yml := `
- foo # comment
- bar
`

	s, err := yamledit.EditBytes("example.yaml", []byte(yml), yamledit.ListAction("$", yamledit.AddValuesToList(0, "zoo")))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
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

	// Add "zoo" to the last position
	s, err := yamledit.EditBytes("example.yaml", []byte(yml), yamledit.ListAction("$", yamledit.AddValuesToList(-1, "zoo")))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
	// Output:
	// - foo # comment
	// - bar
	// - zoo
}

func ExampleAddValuesToList_with_list_bytes() {
	yml := `
- foo # comment
`

	s, err := yamledit.EditBytes("example.yaml", []byte(yml), yamledit.ListAction("$", yamledit.AddValuesToList(-1, yamledit.NewListBytes([]byte(`
- bar # comment 1
- zoo # comment 2
`)))))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
	// Output:
	// - foo # comment
	// - bar # comment 1
	// - zoo # comment 2
}
