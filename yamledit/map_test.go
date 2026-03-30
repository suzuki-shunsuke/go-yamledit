package yamledit_test

import (
	"fmt"
	"log"

	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
)

func ExampleMap() {
	yml := `
name: jack # comment is kept
work: engineer
age: 8
`

	s, err := yamledit.EditBytes([]byte(yml), yamledit.MapAction(
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
	))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
	// Output:
	// name: ryan # comment is kept
	// gender: male
	// job: engineer
}
