package mag_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
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
	type Parent struct {
		Name     string
		Age      int
		Children []Child
	}

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
			&mag.EditMapAction[string, Parent]{
				Edit: func(m *mag.MapValue[string, Parent]) ([]mag.Change, error) {
					mv, ok := m.Map["name"]
					if !ok {
						return nil, nil
					}
					return []mag.Change{
						&mag.ChangeRenameKey{
							Key:  "first_name",
							Node: mv.Node,
						},
					}, nil
				},
			},
		),
		mag.List(
			"$.children",
			// Remove child whose name is "adam
			mag.RemoveItemsFromList[Child](func(value *mag.Node[Child]) (bool, error) {
				return value.Value.Name == "adam", nil
			}),
			// Add a child at index 0
			mag.AddValuesToList(0, map[string]any{"name": "jessica"}),
			// Sort children by name
			mag.SortList[Child](func(a, b *mag.Node[Child]) int {
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
	// first_name: ryan # comment is kept
	// gender: male
	// job: engineer
	// children:
	//   - name: david
	//   - name: jessica
}
