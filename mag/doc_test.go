package mag_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

//nolint:forcetypeassert
func Example() {
	yml := `
name: jack # comment is kept
age: 8
children:
  - name: david
  - name: adam
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	actions := []mag.Action{
		&mag.EditMapValueAction{
			// Edit name to "ryan"
			YAMLPath: "$",
			Match:    mag.MatchMappingValueByKey("name"),
			Edit:     mag.EditMappingValueStatic(mag.NoChange, "ryan"),
		},
		&mag.RemoveKeyAction{
			// Remove the key "age"
			YAMLPath: "$",
			Match:    mag.MatchMappingValueByKey("age"),
		},
		&mag.AddMapKeyAction{
			// Add the key "gender"
			YAMLPath: "$",
			Add:      mag.AddStaticValueToMappingValue("gender", "male", 1),
		},
		&mag.RemoveListItemAction{
			// Remove child whose index is 1
			YAMLPath: "$.children",
			Remove:   mag.RemoveListItemsByIndex(1),
		},
		&mag.AddListItemAction{
			// Add child.
			YAMLPath: "$.children",
			Add:      mag.AddStaticValueToList(map[string]any{"name": "jessica"}, 0),
		},
		&mag.SortListAction{
			// Sort children by name
			YAMLPath: "$.children",
			Sort: func(a, b *mag.SortedItem) int {
				return strings.Compare(
					a.Value.(map[string]any)["name"].(string),
					b.Value.(map[string]any)["name"].(string),
				)
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
