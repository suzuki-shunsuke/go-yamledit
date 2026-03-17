package mag_test

import (
	"fmt"
	"log"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleUpdateMapValueAction_Run() {
	yml := `
name: foo # keep comment
age: 10 # keep comment 2
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	actions := []mag.Action{
		&mag.UpdateMapValueAction{
			YAMLPath: "$",
			Key:      "name",
			Value:    "new name",
		},
		&mag.UpdateMapValueAction{
			YAMLPath: "$",
			Key:      "password", // unknown key
			Value:    "***",
		},
	}
	for _, act := range actions {
		if err := act.Run(file.Docs[0].Body); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(file.String())
	// Output:
	// name: new name # keep comment
	// age: 10 # keep comment 2
}
