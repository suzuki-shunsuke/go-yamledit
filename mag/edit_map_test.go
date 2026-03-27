package mag_test

import (
	"fmt"
	"log"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleEditMapAction_Run() {
	yml := `
name: foo # keep comment
age: 10 # keep comment 2
type: yoo # keep comment 3
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	act := &mag.MapActions{
		YAMLPath: "$",
		Actions: []mag.MapAction{
			&mag.EditMapAction[string]{
				// Change the value of the "name" key to "new name"
				// Match: mag.MatchMappingValueByKey("name"),
				// Edit:  mag.EditMappingValueStatic(mag.NoChange, "new name"),
				Edit: func(m *mag.MapValue[string], _ func(any) error) ([]mag.Change, error) {
					kv, ok := m.Map["name"]
					if !ok {
						return nil, nil
					}
					return []mag.Change{
						&mag.ChangeSetValue{
							Value: "new name",
							Node:  kv.Node,
						},
					}, nil
				},
			},
			&mag.EditMapAction[string]{
				// If the given key does not exist, do nothing
				Edit: func(m *mag.MapValue[string], _ func(any) error) ([]mag.Change, error) {
					kv, ok := m.Map["password"]
					if !ok {
						return nil, nil
					}
					return []mag.Change{
						&mag.ChangeSetValue{
							Value: "***",
							Node:  kv.Node,
						},
					}, nil
				},
			},
			&mag.EditMapAction[string]{
				// Rename the "age" key to "age-2"
				// Match: mag.MatchMappingValueByKey("age"),
				// Edit:  mag.EditMappingValueStatic("age-2", mag.NoChange),
				Edit: func(m *mag.MapValue[string], _ func(any) error) ([]mag.Change, error) {
					kv, ok := m.Map["age"]
					if !ok {
						return nil, nil
					}
					return []mag.Change{
						&mag.ChangeRenameKey{
							Node: kv.Node,
							Key:  "age-2",
						},
					}, nil
				},
			},
			&mag.EditMapAction[string]{
				// Change both key and value
				// key: type => type-2
				// value yoo => yoo-2
				// Match: mag.MatchMappingValueByKey("type"),
				// Edit:  mag.EditMappingValueStatic("type-2", "yoo-2"),
				Edit: func(m *mag.MapValue[string], _ func(any) error) ([]mag.Change, error) {
					kv, ok := m.Map["type"]
					if !ok {
						return nil, nil
					}
					return []mag.Change{
						&mag.ChangeRenameKey{
							Node: kv.Node,
							Key:  "type-2",
						},
						&mag.ChangeSetValue{
							Value: "yoo-2",
							Node:  kv.Node,
						},
					}, nil
				},
			},
		},
	}
	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// name: new name # keep comment
	// age-2: 10 # keep comment 2
	// type-2: yoo-2 # keep comment 3
}
