package mag_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleRemoveListItemAction_Run() {
	yml := `
children:
  - foo # comment
  - bar # comment 2
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	act := &mag.ListActions{
		YAMLPath: "$.children",
		Actions: []mag.ListAction{
			&mag.RemoveListItemAction{
				// Remove the item 0
				Remove: mag.RemoveListItemsByIndex(0),
			},
		},
	}
	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// children:
	//   - bar # comment 2
}

func TestRemoveListItemAction_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		action  mag.ListActions
		want    string
		wantErr bool
	}{
		{
			name: "remove first item",
			yml: `items:
- a
- b
- c
`,
			action: mag.ListActions{
				YAMLPath: "$.items",
				Actions: []mag.ListAction{
					&mag.RemoveListItemAction{
						Remove: mag.RemoveListItemsByIndex(0),
					},
				},
			},
			want: `items:
- b
- c
`,
		},
		{
			name: "remove last item",
			yml: `items:
- a
- b
- c
`,
			action: mag.ListActions{
				YAMLPath: "$.items",
				Actions: []mag.ListAction{
					&mag.RemoveListItemAction{
						Remove: mag.RemoveListItemsByIndex(2),
					},
				},
			},
			want: `items:
- a
- b
`,
		},
		{
			name: "remove middle item",
			yml: `items:
- a
- b
- c
`,
			action: mag.ListActions{
				YAMLPath: "$.items",
				Actions: []mag.ListAction{
					&mag.RemoveListItemAction{
						Remove: mag.RemoveListItemsByIndex(1),
					},
				},
			},
			want: `items:
- a
- c
`,
		},
		{
			name: "nested path",
			yml: `foo:
  items:
  - x
  - y
  - z
`,
			action: mag.ListActions{
				YAMLPath: "$.foo.items",
				Actions: []mag.ListAction{
					&mag.RemoveListItemAction{
						Remove: mag.RemoveListItemsByIndex(1),
					},
				},
			},
			want: `foo:
  items:
  - x
  - z
`,
		},
		{
			name: "with comment preservation",
			yml: `items:
- a # comment1
- b # comment2
- c # comment3
`,
			action: mag.ListActions{
				YAMLPath: "$.items",
				Actions: []mag.ListAction{
					&mag.RemoveListItemAction{
						Remove: mag.RemoveListItemsByIndex(1),
					},
				},
			},
			want: `items:
- a # comment1
- c # comment3
`,
		},
		{
			name: "depth with sequence of sequences",
			yml: `items:
- - a
  - b
- - c
  - d
`,
			action: mag.ListActions{
				YAMLPath: "$.items[*]",
				Actions: []mag.ListAction{
					&mag.RemoveListItemAction{
						Remove: mag.RemoveListItemsByIndex(0),
					},
				},
			},
			want: `items:
- - b
- - d
`,
		},
		{
			name: "invalid yaml path",
			yml: `items:
- a
`,
			action: mag.ListActions{
				YAMLPath: "invalid[",
				Actions: []mag.ListAction{
					&mag.RemoveListItemAction{
						Remove: mag.RemoveListItemsByIndex(0),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Remove is nil",
			yml: `items:
- a
`,
			action: mag.ListActions{
				YAMLPath: "$.items",
				Actions: []mag.ListAction{
					&mag.RemoveListItemAction{},
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
