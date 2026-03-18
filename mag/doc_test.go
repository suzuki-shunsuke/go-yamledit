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
	actions := []mag.Action{
		&mag.MapActions{
			YAMLPath: "$",
			Actions: []mag.MapAction{
				&mag.EditMapValueAction{
					// Edit name to "ryan"
					Match: mag.MatchMappingValueByKey("name"),
					Edit:  mag.EditMappingValueStatic(mag.NoChange, "ryan"),
				},
				// Remove the key "age"
				mag.RemoveKeys("age"),
				// Add the key "gender"
				mag.AddToMap("gender", "male", 1),
			},
		},
		&mag.ListActions{
			YAMLPath: "$.children",
			Actions: []mag.ListAction{
				// Remove child whose index is 1
				mag.RemoveListItemsByIndex(1),
				mag.AddStaticValueToList(map[string]any{"name": "jessica"}, 0),
				&mag.SortListAction[Child]{
					// Sort children by name
					Sort: func(a, b *mag.Node[Child]) int {
						return strings.Compare(a.Value.Name, b.Value.Name)
					},
				},
			},
		},
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
	// children:
	//   - name: david
	//   - name: jessica
}
