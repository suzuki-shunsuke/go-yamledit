package yamledit_test

import (
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
)

type errAction struct{}

func (a errAction) Run(_ ast.Node) error {
	return errors.New("action error")
}

func TestEditBytes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		actions []yamledit.Action
		want    string
		wantErr bool
	}{
		{
			name: "no actions",
			yml: `name: foo
age: 10
`,
			want: `name: foo
age: 10
`,
		},
		{
			name: "single action",
			yml: `name: foo
age: 10
`,
			actions: []yamledit.Action{
				yamledit.MapAction("$", yamledit.SetKey("name", "bar", nil)),
			},
			want: `name: bar
age: 10
`,
		},
		{
			name: "multiple actions",
			yml: `name: foo
age: 10
`,
			actions: []yamledit.Action{
				yamledit.MapAction("$", yamledit.SetKey("name", "bar", nil)),
				yamledit.MapAction("$", yamledit.RemoveKeys("age")),
			},
			want: `name: bar
`,
		},
		{
			name: "comments preserved",
			yml: `name: foo # keep this
age: 10
`,
			actions: []yamledit.Action{
				yamledit.MapAction("$", yamledit.SetKey("name", "bar", nil)),
			},
			want: `name: bar # keep this
age: 10
`,
		},
		{
			name:    "invalid YAML",
			yml:     `{invalid: [`,
			wantErr: true,
		},
		{
			name: "action error",
			yml: `name: foo
`,
			actions: []yamledit.Action{
				errAction{},
			},
			wantErr: true,
		},
		{
			name: "multi-document YAML",
			yml: `name: foo
---
name: bar
`,
			actions: []yamledit.Action{
				yamledit.MapAction("$", yamledit.SetKey("name", "updated", nil)),
			},
			want: `name: updated
---
name: updated
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := yamledit.EditBytes([]byte(tt.yml), tt.actions...)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func ExampleNewBytes() {
	yml := `
- foo # comment
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	act := yamledit.ListAction("$", yamledit.AddValuesToList(0, yamledit.NewBytes([]byte("hello # world"))))
	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// - hello # world
	// - foo # comment
}
