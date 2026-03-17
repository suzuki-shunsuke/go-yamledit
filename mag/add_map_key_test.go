package mag_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleAddMapKeyAction_Run() {
	yml := `
name: foo # keep comment
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	actions := []mag.Action{
		&mag.AddMapKeyAction{
			// Add the key "age" with the value 10
			YAMLPath: "$",
			Add:      mag.NewStaticAddMapKeyEditor("age", 10, 0),
		},
		// If key exist
		// 1. do nothing
		// 2. error
		// 3. overwrite
	}
	for _, act := range actions {
		if err := act.Run(file.Docs[0].Body); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(file.String())
	// Output:
	// age: 10
	// name: foo # keep comment
}

func TestAddMapKeyAction_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		action  mag.AddMapKeyAction
		want    string
		wantErr bool
	}{
		{
			name: "add to beginning",
			yml:  "name: foo\nage: 10\n",
			action: mag.AddMapKeyAction{
				YAMLPath: "$",
				Add:      mag.NewStaticAddMapKeyEditor("first", true, 0),
			},
			want: "first: true\nname: foo\nage: 10\n",
		},
		{
			name: "add to end",
			yml:  "name: foo\nage: 10\n",
			action: mag.AddMapKeyAction{
				YAMLPath: "$",
				Add:      mag.NewStaticAddMapKeyEditor("last", "val", 2),
			},
			want: "name: foo\nage: 10\nlast: val\n",
		},
		{
			name: "add to middle",
			yml:  "a: 1\nb: 2\nc: 3\n",
			action: mag.AddMapKeyAction{
				YAMLPath: "$",
				Add:      mag.NewStaticAddMapKeyEditor("mid", "x", 1),
			},
			want: "a: 1\nmid: x\nb: 2\nc: 3\n",
		},
		{
			name: "negative index",
			yml:  "a: 1\nb: 2\nc: 3\n",
			action: mag.AddMapKeyAction{
				YAMLPath: "$",
				Add:      mag.NewStaticAddMapKeyEditor("neg", "x", -1),
			},
			want: "a: 1\nb: 2\nneg: x\nc: 3\n",
		},
		{
			name: "nested path",
			yml:  "foo:\n  bar: 1\n  baz: 2\n",
			action: mag.AddMapKeyAction{
				YAMLPath: "$.foo",
				Add:      mag.NewStaticAddMapKeyEditor("qux", 99, 0),
			},
			want: "foo:\nqux: 99\n  bar: 1\n  baz: 2\n",
		},
		{
			name: "with comment on value",
			yml:  "name: foo\n",
			action: mag.AddMapKeyAction{
				YAMLPath: "$",
				Add:      mag.NewStaticAddMapKeyEditor("color", mag.WithComment("red", "a nice color"), 0),
			},
			want: "color: red #a nice color\nname: foo\n",
		},
		{
			name: "sequence of mappings",
			yml:  "items:\n- name: a\n  val: 1\n- name: b\n  val: 2\n",
			action: mag.AddMapKeyAction{
				YAMLPath: "$.items",
				Add:      mag.NewStaticAddMapKeyEditor("new", true, 0),
			},
			want: "items:\n- new: true\n    name: a\n    val: 1\n- new: true\n    name: b\n    val: 2\n",
		},
		{
			name: "Add returns ErrNoop",
			yml:  "name: foo\nage: 10\n",
			action: mag.AddMapKeyAction{
				YAMLPath: "$",
				Add: func(_ *ast.MappingNode) (any, any, int, error) {
					return nil, nil, 0, mag.ErrNoop
				},
			},
			want: "name: foo\nage: 10\n",
		},
		{
			name: "invalid yaml path",
			yml:  "name: foo\n",
			action: mag.AddMapKeyAction{
				YAMLPath: "invalid[",
				Add:      mag.NewStaticAddMapKeyEditor("x", "y", 0),
			},
			wantErr: true,
		},
		{
			name: "Add is nil",
			yml:  "name: foo\n",
			action: mag.AddMapKeyAction{
				YAMLPath: "$",
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
