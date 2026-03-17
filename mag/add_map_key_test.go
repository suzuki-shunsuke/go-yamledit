package mag_test

import (
	"fmt"
	"log"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleAddMapKeyAction_Run() {
	yml := `
name: foo # keep comment
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	actions := []mag.Action{
		&mag.AddMapKeyAction{
			// Add the key "age" with the value 10
			YAMLPath: "$",
			Add:      mag.NewStaticAddMapKeyEditor("age", 10, 0),
		},
		// If key exist
		// 1. do nothing
		// 2. error
		// 3. overwrite
	}
	for _, act := range actions {
		if err := act.Run(file.Docs[0].Body); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(file.String())
	// Output:
	// age: 10
	// name: foo # keep comment
}
