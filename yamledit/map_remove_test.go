package yamledit_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
)

func ExampleRemoveKeys() {
	yml := `
name: foo
age: 10 # keep comment
`

	s, err := yamledit.EditBytes("example.yaml", []byte(yml), yamledit.MapAction(
		"$",
		yamledit.RemoveKeys(
			"name",
			"id", // unknown key is ignored
		),
	))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
	// Output:
	// age: 10 # keep comment
}

func TestRemoveKeys(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		action  yamledit.Action
		want    string
		wantErr bool
	}{
		{
			name: "remove root key",
			yml: `name: foo
age: 10
`,
			action: yamledit.MapAction("$", yamledit.RemoveKeys("name")),
			want: `age: 10
`,
		},
		{
			name: "key not found",
			yml: `name: foo
`,
			action: yamledit.MapAction("$", yamledit.RemoveKeys("missing")),
			want: `name: foo
`,
		},
		{
			name: "nested path",
			yml: `foo:
  bar: 1
  baz: 2
`,
			action: yamledit.MapAction("$.foo", yamledit.RemoveKeys("bar")),
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
			action: yamledit.MapAction("$.items", yamledit.RemoveKeys("name")),
			want: `items:
- val: 1
- val: 2
`,
		},
		{
			name: "invalid yaml path",
			yml: `name: foo
`,
			action:  yamledit.MapAction("invalid[", yamledit.RemoveKeys("name")),
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
