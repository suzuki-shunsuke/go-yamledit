package yamledit_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
)

func ExampleRenameKey() {
	yml := `
name: foo
age: 10 # keep comment
`

	s, err := yamledit.EditBytes("example.yaml", []byte(yml), yamledit.MapAction(
		"$",
		yamledit.RenameKey( // Rename name to first_name
			"name",
			"first_name",
			yamledit.Skip,
		),
		yamledit.RenameKey(
			"unknown", // unknown key is ignored
			"unknown-2",
			yamledit.Skip,
		),
	))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
	// Output:
	// first_name: foo
	// age: 10 # keep comment
}

func TestRenameKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		action  yamledit.Action
		want    string
		wantErr bool
	}{
		{
			name: "rename root key",
			yml: `name: foo
age: 10
`,
			action: yamledit.MapAction("$", yamledit.RenameKey("name", "first_name", yamledit.Skip)),
			want: `first_name: foo
age: 10
`,
		},
		{
			name: "key not found",
			yml: `name: foo
`,
			action: yamledit.MapAction("$", yamledit.RenameKey("missing", "new_key", yamledit.Skip)),
			want: `name: foo
`,
		},
		{
			name: "preserve comment",
			yml: `name: foo # important
age: 10
`,
			action: yamledit.MapAction("$", yamledit.RenameKey("name", "first_name", yamledit.Skip)),
			want: `first_name: foo # important
age: 10
`,
		},
		{
			name: "nested path",
			yml: `foo:
  bar: 1
  baz: 2
`,
			action: yamledit.MapAction("$.foo", yamledit.RenameKey("bar", "bar2", yamledit.Skip)),
			want: `foo:
  bar2: 1
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
			action: yamledit.MapAction("$.items[*]", yamledit.RenameKey("val", "value", yamledit.Skip)),
			want: `items:
- name: a
  value: 1
- name: b
  value: 2
`,
		},
		{
			name: "same key noop",
			yml: `name: foo
`,
			action: yamledit.MapAction("$", yamledit.RenameKey("name", "name", yamledit.Skip)),
			want: `name: foo
`,
		},
		{
			name: "skip when duplicate",
			yml: `name: foo
first_name: bar
`,
			action: yamledit.MapAction("$", yamledit.RenameKey("name", "first_name", yamledit.Skip)),
			want: `name: foo
first_name: bar
`,
		},
		{
			name: "ignore existing key",
			yml: `name: foo
first_name: bar
`,
			action: yamledit.MapAction("$", yamledit.RenameKey("name", "first_name", yamledit.IgnoreExistingKey)),
			want: `first_name: foo
`,
		},
		{
			name: "remove old key",
			yml: `name: foo
first_name: bar
`,
			action: yamledit.MapAction("$", yamledit.RenameKey("name", "first_name", yamledit.RemoveOldKey)),
			want: `first_name: bar
`,
		},
		{
			name: "raise error on duplicate",
			yml: `name: foo
first_name: bar
`,
			action:  yamledit.MapAction("$", yamledit.RenameKey("name", "first_name", yamledit.RaiseError)),
			wantErr: true,
		},
		{
			name:    "invalid yaml path",
			yml:     `name: foo`,
			action:  yamledit.MapAction("invalid[", yamledit.RenameKey("name", "new_name", yamledit.Skip)),
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
