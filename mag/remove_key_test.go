package mag_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleRemoveKeyAction_Run() {
	yml := `
name: foo
age: 10 # keep comment
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	act := &mag.MapActions{
		YAMLPath: "$",
		Actions: []mag.MapAction{
			&mag.RemoveKeyAction{
				Match: mag.MatchMappingValueByKey("name"),
			},
			&mag.RemoveKeyAction{
				Match: mag.MatchMappingValueByKey("id"), // unknown key
			},
		},
	}
	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// age: 10 # keep comment
}

func TestRemoveKeyAction_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		action  mag.MapActions
		want    string
		wantErr bool
	}{
		{
			name: "remove root key",
			yml: `name: foo
age: 10
`,
			action: mag.MapActions{
				YAMLPath: "$",
				Actions: []mag.MapAction{
					&mag.RemoveKeyAction{
						Match: mag.MatchMappingValueByKey("name"),
					},
				},
			},
			want: `age: 10
`,
		},
		{
			name: "key not found",
			yml: `name: foo
`,
			action: mag.MapActions{
				YAMLPath: "$",
				Actions: []mag.MapAction{
					&mag.RemoveKeyAction{
						Match: mag.MatchMappingValueByKey("missing"),
					},
				},
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
			action: mag.MapActions{
				YAMLPath: "$.foo",
				Actions: []mag.MapAction{
					&mag.RemoveKeyAction{
						Match: mag.MatchMappingValueByKey("bar"),
					},
				},
			},
			want: `foo:
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
			action: mag.MapActions{
				YAMLPath: "$.items",
				Actions: []mag.MapAction{
					&mag.RemoveKeyAction{
						Match: mag.MatchMappingValueByKey("name"),
					},
				},
			},
			want: `items:
- val: 1
- val: 2
`,
		},
		{
			name: "invalid yaml path",
			yml: `name: foo
`,
			action: mag.MapActions{
				YAMLPath: "invalid[",
				Actions: []mag.MapAction{
					&mag.RemoveKeyAction{
						Match: mag.MatchMappingValueByKey("name"),
					},
				},
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
