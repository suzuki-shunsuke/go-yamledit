package mag_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleEditMapValueAction_Run() {
	yml := `
name: foo # keep comment
age: 10 # keep comment 2
type: yoo # keep comment 3
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	actions := []mag.Action{
		&mag.EditMapValueAction{
			// Change the value of the "name" key to "new name"
			YAMLPath: "$",
			Match:    mag.MatchMappingValueByKey("name"),
			Edit:     mag.EditMappingValueStatic(mag.NoChange, "new name"),
		},
		&mag.EditMapValueAction{
			// If the given key does not exist, do nothing
			YAMLPath: "$",
			Match:    mag.MatchMappingValueByKey("password"), // unknown key
			Edit:     mag.EditMappingValueStatic(mag.NoChange, "***"),
		},
		&mag.EditMapValueAction{
			// Rename the "age" key to "age-2"
			YAMLPath: "$",
			Match:    mag.MatchMappingValueByKey("age"),
			Edit:     mag.EditMappingValueStatic("age-2", mag.NoChange),
		},
		&mag.EditMapValueAction{
			// Change both key and value
			// key: type => type-2
			// value yoo => yoo-2
			YAMLPath: "$",
			Match:    mag.MatchMappingValueByKey("type"),
			Edit:     mag.EditMappingValueStatic("type-2", "yoo-2"),
		},
	}
	for _, act := range actions {
		if err := act.Run(file.Docs[0].Body); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(file.String())
	// Output:
	// name: new name # keep comment
	// age-2: 10 # keep comment 2
	// type-2: yoo-2 # keep comment 3
}

func TestEditMapValueAction_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		action  mag.EditMapValueAction
		want    string
		wantErr bool
	}{
		{
			name: "update root value",
			yml: `name: foo
age: 10
`,
			action: mag.EditMapValueAction{
				YAMLPath: "$",
				Match:    mag.MatchMappingValueByKey("name"),
				Edit:     mag.EditMappingValueStatic(mag.NoChange, "bar"),
			},
			want: `name: bar
age: 10
`,
		},
		{
			name: "key not found",
			yml: `name: foo
`,
			action: mag.EditMapValueAction{
				YAMLPath: "$",
				Match:    mag.MatchMappingValueByKey("missing"),
				Edit:     mag.EditMappingValueStatic(mag.NoChange, "val"),
			},
			want: `name: foo
`,
		},
		{
			name: "nested path",
			yml: `foo:
  bar: 1
  baz: 2
`,
			action: mag.EditMapValueAction{
				YAMLPath: "$.foo",
				Match:    mag.MatchMappingValueByKey("bar"),
				Edit:     mag.EditMappingValueStatic(mag.NoChange, 99),
			},
			want: `foo:
  bar: 99
  baz: 2
`,
		},
		{
			name: "sequence of mappings",
			yml: `items:
- name: a
  val: 1
- name: b
  val: 2
`,
			action: mag.EditMapValueAction{
				YAMLPath: "$.items",
				Match:    mag.MatchMappingValueByKey("val"),
				Edit:     mag.EditMappingValueStatic(mag.NoChange, 100),
			},
			want: `items:
- name: a
  val: 100
- name: b
  val: 100
`,
		},
		{
			name: "invalid yaml path",
			yml: `name: foo
`,
			action: mag.EditMapValueAction{
				YAMLPath: "invalid[",
				Match:    mag.MatchMappingValueByKey("name"),
				Edit:     mag.EditMappingValueStatic(mag.NoChange, "bar"),
			},
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
