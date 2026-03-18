package mag_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleAddToMap() {
	yml := `
name: foo # keep comment
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	// Add the key "age" with the value 10
	act := mag.Map("$", mag.AddToMap("age", 10, 0))
	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
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
		action  mag.Action
		want    string
		wantErr bool
	}{
		{
			name: "add to beginning",
			yml: `name: foo
age: 10
`,
			action: mag.Map("$", mag.AddToMap("first", true, 0)),
			want: `first: true
name: foo
age: 10
`,
		},
		{
			name: "add to end",
			yml: `name: foo
age: 10
`,
			action: mag.Map("$", mag.AddToMap("last", "val", 2)),
			want: `name: foo
age: 10
last: val
`,
		},
		{
			name: "add to middle",
			yml: `a: 1
b: 2
c: 3
`,
			action: mag.Map("$", mag.AddToMap("mid", "x", 1)),
			want: `a: 1
mid: x
b: 2
c: 3
`,
		},
		{
			name: "negative index",
			yml: `a: 1
b: 2
c: 3
`,
			action: mag.Map("$", mag.AddToMap("neg", "x", -1)),
			want: `a: 1
b: 2
neg: x
c: 3
`,
		},
		{
			name: "nested path",
			yml: `foo:
  bar: 1
  baz: 2
`,
			action: mag.Map("$.foo", mag.AddToMap("qux", 99, 0)),
			want: `foo:
qux: 99
  bar: 1
  baz: 2
`,
		},
		{
			name: "with comment on value",
			yml: `name: foo
`,
			action: mag.Map("$", mag.AddToMap("color", mag.WithComment("red", "a nice color"), 0)),
			want: `color: red #a nice color
name: foo
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
			action: mag.Map("$.items", mag.AddToMap("new", true, 0)),
			want: `items:
- new: true
    name: a
    val: 1
- new: true
    name: b
    val: 2
`,
		},
		{
			name: "Add returns ErrNoop",
			yml: `name: foo
age: 10
`,
			action: mag.Map(
				"$",
				mag.AddToMapByFunc(func(_ *ast.MappingNode) (any, any, int, error) {
					return nil, nil, 0, mag.ErrNoop
				}),
			),
			want: `name: foo
age: 10
`,
		},
		{
			name: "invalid yaml path",
			yml: `name: foo
`,
			action:  mag.Map("invalid[", mag.AddToMap("x", "y", 0)),
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
