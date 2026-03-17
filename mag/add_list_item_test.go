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
			// Add the key "age" with the value 10
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
			yml:  "items:\n- a\n- b\n",
			action: mag.AddListItemAction{
				YAMLPath: "$.items",
				Add:      mag.NewStaticAddListItemEditor("first", 0),
			},
			want: "items:\n- first\n- a\n- b\n",
		},
		{
			name: "add to end",
			yml:  "items:\n- a\n- b\n",
			action: mag.AddListItemAction{
				YAMLPath: "$.items",
				Add:      mag.NewStaticAddListItemEditor("last", 2),
			},
			want: "items:\n- a\n- b\n- last\n",
		},
		{
			name: "add to middle",
			yml:  "items:\n- a\n- b\n- c\n",
			action: mag.AddListItemAction{
				YAMLPath: "$.items",
				Add:      mag.NewStaticAddListItemEditor("mid", 1),
			},
			want: "items:\n- a\n- mid\n- b\n- c\n",
		},
		{
			name: "nested path",
			yml:  "foo:\n  items:\n  - x\n  - y\n",
			action: mag.AddListItemAction{
				YAMLPath: "$.foo.items",
				Add:      mag.NewStaticAddListItemEditor("z", 0),
			},
			want: "foo:\n  items:\n  - z\n  - x\n  - y\n",
		},
		{
			name: "with comment preservation",
			yml:  "items:\n- a # comment1\n- b # comment2\n",
			action: mag.AddListItemAction{
				YAMLPath: "$.items",
				Add:      mag.NewStaticAddListItemEditor("new", 1),
			},
			want: "items:\n- a # comment1\n- new\n- b # comment2\n",
		},
		{
			name: "Add returns Noop",
			yml:  "items:\n- a\n- b\n",
			action: mag.AddListItemAction{
				YAMLPath: "$.items",
				Add: func(_ *ast.SequenceNode) (any, int, error) {
					return mag.Noop, 0, nil
				},
			},
			want: "items:\n- a\n- b\n",
		},
		{
			name: "invalid yaml path",
			yml:  "items:\n- a\n",
			action: mag.AddListItemAction{
				YAMLPath: "invalid[",
				Add:      mag.NewStaticAddListItemEditor("x", 0),
			},
			wantErr: true,
		},
		{
			name: "Add is nil",
			yml:  "items:\n- a\n",
			action: mag.AddListItemAction{
				YAMLPath: "$.items",
			},
			wantErr: true,
		},
		{
			name: "negative depth",
			yml:  "items:\n- a\n",
			action: mag.AddListItemAction{
				YAMLPath: "$.items",
				Add:      mag.NewStaticAddListItemEditor("x", 0),
				Depth:    -1,
			},
			wantErr: true,
		},
		{
			name: "depth with sequence of sequences",
			yml:  "items:\n- - a\n  - b\n- - c\n  - d\n",
			action: mag.AddListItemAction{
				YAMLPath: "$.items",
				Add:      mag.NewStaticAddListItemEditor("new", 0),
				Depth:    1,
			},
			want: "items:\n- - new\n  - a\n  - b\n- - new\n  - c\n  - d\n",
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
