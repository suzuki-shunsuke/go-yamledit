package mag_test

import (
	"fmt"
	"log"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleWithComment() {
	yml := `
- foo
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	act := &mag.AddListItemAction{
		// Add "zoo" with comment
		YAMLPath: "$",
		Add:      mag.NewStaticAddListItemEditor(mag.WithComment("zoo", " comment is added"), 1),
	}
	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// - foo
	// - zoo # comment is added
}
