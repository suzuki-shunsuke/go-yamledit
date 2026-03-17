package mag_test

import (
	"fmt"
	"log"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleRenameKeyAction_Run() {
	yml := `
name: foo # keep comment
age: 10 # keep comment 2
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	actions := []mag.Action{
		&mag.RenameKeyAction{
			YAMLPath: "$",
			OldKey:   "name",
			NewKey:   "id",
		},
		&mag.RenameKeyAction{
			YAMLPath: "$",
			OldKey:   "password", // unknown key
			NewKey:   "passwd",
		},
	}
	for _, act := range actions {
		if err := act.Run(file.Docs[0].Body); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(file.String())
	// Output:
	// id: foo # keep comment
	// age: 10 # keep comment 2
}
