package yamledit_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
)

func ExampleSetKey() {
	yml := `
name: foo # keep comment
age: 10
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	act := yamledit.MapAction(
		"$",
		// Edit name to "ryan"
		yamledit.SetKey("name", "ryan", nil),
		yamledit.SetKey("gender", "male", &yamledit.SetKeyOption{
			InsertLocations: []*yamledit.InsertLocation{
				{
					AfterKey: "foo", // Ignore unknown key
				},
				{
					BeforeKey: "age",
				},
			},
		}),
	)

	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// name: ryan # keep comment
	// gender: male
	// age: 10
}

func TestSetKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		action  yamledit.Action
		want    string
		wantErr bool
	}{
		{
			name: "set existing key",
			yml: `name: foo
age: 10
`,
			action: yamledit.MapAction("$", yamledit.SetKey("name", "bar", nil)),
			want: `name: bar
age: 10
`,
		},
		{
			name: "add new key append",
			yml: `name: foo
`,
			action: yamledit.MapAction("$", yamledit.SetKey("age", 10, nil)),
			want: `name: foo
age: 10
`,
		},
		{
			name: "ignore if key not exist",
			yml: `name: foo
`,
			action: yamledit.MapAction("$", yamledit.SetKey("age", 10, &yamledit.SetKeyOption{
				IgnoreIfKeyNotExist: true,
			})),
			want: `name: foo
`,
		},
		{
			name: "ignore if key exist",
			yml: `name: foo
`,
			action: yamledit.MapAction("$", yamledit.SetKey("name", "bar", &yamledit.SetKeyOption{
				IgnoreIfKeyExist: true,
			})),
			want: `name: foo
`,
		},
		{
			name: "insert first",
			yml: `name: foo
age: 10
`,
			action: yamledit.MapAction("$", yamledit.SetKey("gender", "male", &yamledit.SetKeyOption{
				InsertLocations: []*yamledit.InsertLocation{
					{First: true},
				},
			})),
			want: `gender: male
name: foo
age: 10
`,
		},
		{
			name: "insert before key",
			yml: `name: foo
age: 10
`,
			action: yamledit.MapAction("$", yamledit.SetKey("gender", "male", &yamledit.SetKeyOption{
				InsertLocations: []*yamledit.InsertLocation{
					{BeforeKey: "age"},
				},
			})),
			want: `name: foo
gender: male
age: 10
`,
		},
		{
			name: "insert after key",
			yml: `name: foo
age: 10
`,
			action: yamledit.MapAction("$", yamledit.SetKey("gender", "male", &yamledit.SetKeyOption{
				InsertLocations: []*yamledit.InsertLocation{
					{AfterKey: "name"},
				},
			})),
			want: `name: foo
gender: male
age: 10
`,
		},
		{
			name: "insert location fallback to append",
			yml: `name: foo
age: 10
`,
			action: yamledit.MapAction("$", yamledit.SetKey("gender", "male", &yamledit.SetKeyOption{
				InsertLocations: []*yamledit.InsertLocation{
					{BeforeKey: "missing"},
				},
			})),
			want: `name: foo
age: 10
gender: male
`,
		},
		{
			name: "preserve comment on update",
			yml: `name: foo # important
age: 10
`,
			action: yamledit.MapAction("$", yamledit.SetKey("name", "bar", nil)),
			want: `name: bar # important
age: 10
`,
		},
		{
			name: "nested path",
			yml: `foo:
  bar: 1
  baz: 2
`,
			action: yamledit.MapAction("$.foo", yamledit.SetKey("bar", 99, nil)),
			want: `foo:
  bar: 99
  baz: 2
`,
		},
		{
			name: "add new key nested path",
			yml: `foo:
  bar: 1
`,
			action: yamledit.MapAction("$.foo", yamledit.SetKey("baz", 2, nil)),
			want: `foo:
  bar: 1
  baz: 2
`,
		},
		{
			name: "set key with map value",
			yml: `aliases: {}
`,
			action: yamledit.MapAction("$", yamledit.SetKey("aliases", map[string]string{"my-rule": "https://example.com"}, nil)),
			want: `aliases:
  my-rule: https://example.com
`,
		},
		{
			name: "set key with map value nested",
			yml: `foo:
  aliases: {}
`,
			action: yamledit.MapAction("$.foo", yamledit.SetKey("aliases", map[string]string{"my-rule": "https://example.com"}, nil)),
			want: `foo:
  aliases:
    my-rule: https://example.com
`,
		},
		{
			name: "clear comment on update",
			yml: `name: foo # important
age: 10
`,
			action: yamledit.MapAction("$", yamledit.SetKey("name", "bar", &yamledit.SetKeyOption{
				ClearComment: true,
			})),
			want: `name: bar
age: 10
`,
		},
		{
			name: "clear comment false preserves comment",
			yml: `name: foo # important
`,
			action: yamledit.MapAction("$", yamledit.SetKey("name", "bar", &yamledit.SetKeyOption{
				ClearComment: false,
			})),
			want: `name: bar # important
`,
		},
		{
			name: "clear comment on value without comment",
			yml: `name: foo
`,
			action: yamledit.MapAction("$", yamledit.SetKey("name", "bar", &yamledit.SetKeyOption{
				ClearComment: true,
			})),
			want: `name: bar
`,
		},
		{
			name:    "invalid yaml path",
			yml:     `name: foo`,
			action:  yamledit.MapAction("invalid[", yamledit.SetKey("name", "bar", nil)),
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
