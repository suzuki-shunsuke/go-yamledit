package mag_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleRemoveKeys() {
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
		mag.RemoveKeys(
			"name",
			"id", // unknown key is ignored
		),
	)

	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// age: 10 # keep comment
}

func TestRemoveKeys(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		action  mag.Action
		want    string
		wantErr bool
	}{
		{
			name: "remove root key",
			yml: `name: foo
age: 10
`,
			action: mag.MapAction("$", mag.RemoveKeys("name")),
			want: `age: 10
`,
		},
		{
			name: "key not found",
			yml: `name: foo
`,
			action: mag.MapAction("$", mag.RemoveKeys("missing")),
			want: `name: foo
`,
		},
		{
			name: "nested path",
			yml: `foo:
  bar: 1
  baz: 2
`,
			action: mag.MapAction("$.foo", mag.RemoveKeys("bar")),
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
			action: mag.MapAction("$.items", mag.RemoveKeys("name")),
			want: `items:
- val: 1
- val: 2
`,
		},
		{
			name: "invalid yaml path",
			yml: `name: foo
`,
			action:  mag.MapAction("invalid[", mag.RemoveKeys("name")),
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
