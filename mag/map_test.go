package mag_test

import (
	"fmt"
	"log"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
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
	actions := []mag.Action{
		mag.Map(
			"$",
			// Edit name to "ryan"
			mag.SetKey("name", "ryan", nil),
			// Remove the key "age"
			mag.RemoveKeys("age"),
			// Rename the key "work" to "job"
			mag.RenameKey("work", "job"),
			// Add the key "gender" after "name"
			mag.SetKey("gender", "male", &mag.SetKeyOption{
				InsertLocations: []*mag.InsertLocation{
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
