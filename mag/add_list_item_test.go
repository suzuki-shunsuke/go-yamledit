package mag_test

import (
	"fmt"
	"log"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleAddListItemAction_Run() {
	yml := `
children:
  - foo # comment
  - bar
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	actions := []mag.Action{
		&mag.AddListItemAction{
			// Add the key "age" with the value 10
			YAMLPath: "$.children",
			Add:      mag.NewStaticAddListItemEditor("zoo", 0),
		},
	}
	for _, act := range actions {
		if err := act.Run(file.Docs[0].Body); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(file.String())
	// Output:
	// children:
	//   - zoo
	//   - foo # comment
	//   - bar
}
