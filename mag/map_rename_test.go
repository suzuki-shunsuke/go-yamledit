package mag_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleRenameKey() {
	yml := `
name: foo
age: 10 # keep comment
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	act := mag.MapAction(
		"$",
		mag.RenameKey( // Rename name to first_name
			"name",
			"first_name",
			mag.Skip,
		),
		mag.RenameKey(
			"unknown", // unknown key is ignored
			"unknown-2",
			mag.Skip,
		),
	)

	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// first_name: foo
	// age: 10 # keep comment
}

func TestRenameKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		action  mag.Action
		want    string
		wantErr bool
	}{
		{
			name: "rename root key",
			yml: `name: foo
age: 10
`,
			action: mag.MapAction("$", mag.RenameKey("name", "first_name", mag.Skip)),
			want: `first_name: foo
age: 10
`,
		},
		{
			name: "key not found",
			yml: `name: foo
`,
			action: mag.MapAction("$", mag.RenameKey("missing", "new_key", mag.Skip)),
			want: `name: foo
`,
		},
		{
			name: "preserve comment",
			yml: `name: foo # important
age: 10
`,
			action: mag.MapAction("$", mag.RenameKey("name", "first_name", mag.Skip)),
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
			action: mag.MapAction("$.foo", mag.RenameKey("bar", "bar2", mag.Skip)),
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
			action: mag.MapAction("$.items[*]", mag.RenameKey("val", "value", mag.Skip)),
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
			action: mag.MapAction("$", mag.RenameKey("name", "name", mag.Skip)),
			want: `name: foo
`,
		},
		{
			name: "skip when duplicate",
			yml: `name: foo
first_name: bar
`,
			action: mag.MapAction("$", mag.RenameKey("name", "first_name", mag.Skip)),
			want: `name: foo
first_name: bar
`,
		},
		{
			name: "ignore existing key",
			yml: `name: foo
first_name: bar
`,
			action: mag.MapAction("$", mag.RenameKey("name", "first_name", mag.IgnoreExistingKey)),
			want: `first_name: foo
`,
		},
		{
			name: "remove old key",
			yml: `name: foo
first_name: bar
`,
			action: mag.MapAction("$", mag.RenameKey("name", "first_name", mag.RemoveOldKey)),
			want: `first_name: bar
`,
		},
		{
			name: "raise error on duplicate",
			yml: `name: foo
first_name: bar
`,
			action:  mag.MapAction("$", mag.RenameKey("name", "first_name", mag.RaiseError)),
			wantErr: true,
		},
		{
			name:    "invalid yaml path",
			yml:     `name: foo`,
			action:  mag.MapAction("invalid[", mag.RenameKey("name", "new_name", mag.Skip)),
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
