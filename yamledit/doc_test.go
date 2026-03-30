package yamledit_test

import (
	"fmt"
	"log"
	"strings"

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

	s, err := yamledit.EditBytes([]byte(yml),
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
		))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
	// Output:
	// name: ryan # comment is kept
	// gender: male
	// job: engineer
	// children:
	//   - name: david
	//   - name: jessica
}
