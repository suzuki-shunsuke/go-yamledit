package mag_test

import (
	"fmt"
	"log"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleRemoveListItemAction_Run() {
	yml := `
children:
  - foo # comment
  - bar # comment 2
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	actions := []mag.Action{
		&mag.RemoveListItemAction{
			// Remove the item 0
			YAMLPath: "$.children",
			Remove:   mag.NewStaticRemoveListItemEditor(0),
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
	//   - bar # comment 2
}
