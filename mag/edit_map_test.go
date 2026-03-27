package mag_test

import (
	"errors"
	"fmt"
	"log"
	"testing"

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
			&mag.EditMapAction[string, any]{
				// Change the value of the "name" key to "new name"
				// Match: mag.MatchMappingValueByKey("name"),
				// Edit:  mag.EditMappingValueStatic(mag.NoChange, "new name"),
				Edit: func(m *mag.MapValue[string, any]) ([]mag.Change, error) {
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
			&mag.EditMapAction[string, any]{
				// If the given key does not exist, do nothing
				Edit: func(m *mag.MapValue[string, any]) ([]mag.Change, error) {
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
			&mag.EditMapAction[string, any]{
				// Rename the "age" key to "age-2"
				// Match: mag.MatchMappingValueByKey("age"),
				// Edit:  mag.EditMappingValueStatic("age-2", mag.NoChange),
				Edit: func(m *mag.MapValue[string, any]) ([]mag.Change, error) {
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
			&mag.EditMapAction[string, any]{
				// Change both key and value
				// key: type => type-2
				// value yoo => yoo-2
				// Match: mag.MatchMappingValueByKey("type"),
				// Edit:  mag.EditMappingValueStatic("type-2", "yoo-2"),
				Edit: func(m *mag.MapValue[string, any]) ([]mag.Change, error) {
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

func TestEditMapAction_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		action  mag.Action
		want    string
		wantErr bool
	}{
		{
			name: "set value",
			yml: `name: foo
age: 10
`,
			action: mag.Map("$", &mag.EditMapAction[string, any]{
				Edit: func(m *mag.MapValue[string, any]) ([]mag.Change, error) {
					kv, ok := m.Map["name"]
					if !ok {
						return nil, nil
					}
					return []mag.Change{
						&mag.ChangeSetValue{
							Value: "bar",
							Node:  kv.Node,
						},
					}, nil
				},
			}),
			want: `name: bar
age: 10
`,
		},
		{
			name: "rename key",
			yml: `name: foo
age: 10
`,
			action: mag.Map("$", &mag.EditMapAction[string, any]{
				Edit: func(m *mag.MapValue[string, any]) ([]mag.Change, error) {
					kv, ok := m.Map["name"]
					if !ok {
						return nil, nil
					}
					return []mag.Change{
						&mag.ChangeRenameKey{
							Key:  "first_name",
							Node: kv.Node,
						},
					}, nil
				},
			}),
			want: `first_name: foo
age: 10
`,
		},
		{
			name: "key not found",
			yml: `name: foo
`,
			action: mag.Map("$", &mag.EditMapAction[string, any]{
				Edit: func(m *mag.MapValue[string, any]) ([]mag.Change, error) {
					_, ok := m.Map["missing"]
					if !ok {
						return nil, nil
					}
					return nil, nil
				},
			}),
			want: `name: foo
`,
		},
		{
			name: "preserve comment",
			yml: `name: foo # important
`,
			action: mag.Map("$", &mag.EditMapAction[string, any]{
				Edit: func(m *mag.MapValue[string, any]) ([]mag.Change, error) {
					kv := m.Map["name"]
					return []mag.Change{
						&mag.ChangeSetValue{
							Value: "bar",
							Node:  kv.Node,
						},
					}, nil
				},
			}),
			want: `name: bar # important
`,
		},
		{
			name: "edit func returns error",
			yml: `name: foo
`,
			action: mag.Map("$", &mag.EditMapAction[string, any]{
				Edit: func(_ *mag.MapValue[string, any]) ([]mag.Change, error) {
					return nil, errors.New("edit error")
				},
			}),
			wantErr: true,
		},
		{
			name: "multiple changes",
			yml: `name: foo
age: 10
`,
			action: mag.Map("$", &mag.EditMapAction[string, any]{
				Edit: func(m *mag.MapValue[string, any]) ([]mag.Change, error) {
					return []mag.Change{
						&mag.ChangeSetValue{
							Value: "bar",
							Node:  m.Map["name"].Node,
						},
						&mag.ChangeSetValue{
							Value: 20,
							Node:  m.Map["age"].Node,
						},
					}, nil
				},
			}),
			want: `name: bar
age: 20
`,
		},
		{
			name: "no changes",
			yml: `name: foo
`,
			action: mag.Map("$", &mag.EditMapAction[string, any]{
				Edit: func(_ *mag.MapValue[string, any]) ([]mag.Change, error) {
					return nil, nil
				},
			}),
			want: `name: foo
`,
		},
		{
			name: "nested path",
			yml: `foo:
  bar: 1
  baz: 2
`,
			action: mag.Map("$.foo", &mag.EditMapAction[string, any]{
				Edit: func(m *mag.MapValue[string, any]) ([]mag.Change, error) {
					kv := m.Map["bar"]
					return []mag.Change{
						&mag.ChangeSetValue{
							Value: 99,
							Node:  kv.Node,
						},
					}, nil
				},
			}),
			want: `foo:
  bar: 99
  baz: 2
`,
		},
		{
			name: "invalid yaml path",
			yml:  `name: foo`,
			action: mag.Map("invalid[", &mag.EditMapAction[string, any]{
				Edit: func(_ *mag.MapValue[string, any]) ([]mag.Change, error) {
					return nil, nil
				},
			}),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			file, err := parser.ParseBytes([]byte(tt.yml), parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			err = tt.action.Run(file.Docs[0].Body)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			got := file.String()
			if got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}
