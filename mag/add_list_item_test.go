package mag_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleAddListItemAction_Run() {
	yml := `
children:
  - foo # comment
  - bar
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	actions := []mag.Action{
		&mag.AddListItemAction{
			// Add "zoo" to the first position
			YAMLPath: "$.children",
			Add:      mag.NewStaticAddListItemEditor("zoo", 0),
		},
	}
	for _, act := range actions {
		if err := act.Run(file.Docs[0].Body); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(file.String())
	// Output:
	// children:
	//   - zoo
	//   - foo # comment
	//   - bar
}

func ExampleAddListItemAction_Run_negative_index() {
	yml := `
- foo # comment
- bar
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	act := &mag.AddListItemAction{
		// Add "zoo" to the last position
		YAMLPath: "$",
		Add:      mag.NewStaticAddListItemEditor("zoo", -1),
	}
	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// - foo # comment
	// - bar
	// - zoo
}

func TestAddListItemAction_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		action  mag.AddListItemAction
		want    string
		wantErr bool
	}{
		{
			name: "add to beginning",
			yml: `items:
- a
- b
`,
			action: mag.AddListItemAction{
				YAMLPath: "$.items",
				Add:      mag.NewStaticAddListItemEditor("first", 0),
			},
			want: `items:
- first
- a
- b
`,
		},
		{
			name: "add to end",
			yml: `items:
- a
- b
`,
			action: mag.AddListItemAction{
				YAMLPath: "$.items",
				Add:      mag.NewStaticAddListItemEditor("last", 2),
			},
			want: `items:
- a
- b
- last
`,
		},
		{
			name: "add to middle",
			yml: `items:
- a
- b
- c
`,
			action: mag.AddListItemAction{
				YAMLPath: "$.items",
				Add:      mag.NewStaticAddListItemEditor("mid", 1),
			},
			want: `items:
- a
- mid
- b
- c
`,
		},
		{
			name: "nested path",
			yml: `foo:
  items:
  - x
  - y
`,
			action: mag.AddListItemAction{
				YAMLPath: "$.foo.items",
				Add:      mag.NewStaticAddListItemEditor("z", 0),
			},
			want: `foo:
  items:
  - z
  - x
  - y
`,
		},
		{
			name: "with comment preservation",
			yml: `items:
- a # comment1
- b # comment2
`,
			action: mag.AddListItemAction{
				YAMLPath: "$.items",
				Add:      mag.NewStaticAddListItemEditor("new", 1),
			},
			want: `items:
- a # comment1
- new
- b # comment2
`,
		},
		{
			name: "Add returns ErrNoop",
			yml: `items:
- a
- b
`,
			action: mag.AddListItemAction{
				YAMLPath: "$.items",
				Add: func(_ *ast.SequenceNode) (any, int, error) {
					return nil, 0, mag.ErrNoop
				},
			},
			want: `items:
- a
- b
`,
		},
		{
			name: "invalid yaml path",
			yml: `items:
- a
`,
			action: mag.AddListItemAction{
				YAMLPath: "invalid[",
				Add:      mag.NewStaticAddListItemEditor("x", 0),
			},
			wantErr: true,
		},
		{
			name: "Add is nil",
			yml: `items:
- a
`,
			action: mag.AddListItemAction{
				YAMLPath: "$.items",
			},
			wantErr: true,
		},
		{
			name: "depth with sequence of sequences",
			yml: `items:
- - a
  - b
- - c
  - d
`,
			action: mag.AddListItemAction{
				YAMLPath: "$.items[*]",
				Add:      mag.NewStaticAddListItemEditor("new", 0),
			},
			want: `items:
- - new
  - a
  - b
- - new
  - c
  - d
`,
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
