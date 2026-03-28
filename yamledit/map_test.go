package yamledit_test

import (
	"fmt"
	"log"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
)

func ExampleMap() {
	yml := `
name: jack # comment is kept
work: engineer
age: 8
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	actions := []yamledit.Action{
		yamledit.MapAction(
			"$",
			// Edit name to "ryan"
			yamledit.SetKey("name", "ryan", nil),
			// Remove the key "age"
			yamledit.RemoveKeys("age"),
			// Rename the key "work" to "job"
			yamledit.RenameKey("work", "job", yamledit.Skip),
			// Add the key "gender" after "name"
			yamledit.SetKey("gender", "male", &yamledit.SetKeyOption{
				InsertLocations: []*yamledit.InsertLocation{
					{
						AfterKey: "name",
					},
				},
			}),
		),
	}
	for _, act := range actions {
		if err := act.Run(file.Docs[0].Body); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(file.String())
	// Output:
	// name: ryan # comment is kept
	// gender: male
	// job: engineer
}
