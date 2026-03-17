package mag_test

import (
	"fmt"
	"log"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleRemoveKeyAction_Run() {
	yml := `
name: foo
age: 10 # keep comment
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	actions := []mag.Action{
		&mag.RemoveKeyAction{
			YAMLPath: "$",
			Key:      "name",
		},
		&mag.RemoveKeyAction{
			YAMLPath: "$",
			Key:      "id", // unknown key
		},
	}
	for _, act := range actions {
		if err := act.Run(file.Docs[0].Body); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(file.String())
	// Output:
	// age: 10 # keep comment
}
