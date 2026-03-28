package yamledit_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
)

func Example() {
	yml := `
name: jack # comment is kept
work: engineer
age: 8
children:
  - name: david
  - name: adam
`

	type Child struct {
		Name string
	}

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
		yamledit.ListAction(
			"$.children",
			// Remove child whose name is "adam
			yamledit.RemoveValuesFromList[Child](func(value *yamledit.Node[Child]) (bool, error) {
				return value.Value.Name == "adam", nil
			}),
			// Add a child at index 0
			yamledit.AddValuesToList(0, Child{
				Name: "jessica",
			}),
			// Sort children by name
			yamledit.SortList[Child](func(a, b *yamledit.Node[Child]) int {
				return strings.Compare(a.Value.Name, b.Value.Name)
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
	// children:
	//   - name: david
	//   - name: jessica
}
