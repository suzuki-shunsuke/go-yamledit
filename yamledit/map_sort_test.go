package yamledit_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
)

func ascKeyFunc(a, b *yamledit.KeyValue[string]) int {
	if a.Key < b.Key {
		return -1
	}
	if a.Key > b.Key {
		return 1
	}
	return 0
}

func descKeyFunc(a, b *yamledit.KeyValue[string]) int {
	return -ascKeyFunc(a, b)
}

func ExampleSortKey() {
	yml := `
name: foo # keep comment
age: 10
job: engineer
`

	s, err := yamledit.EditBytes("example.yaml", []byte(yml), yamledit.MapAction(
		"$",
		yamledit.SortKey(func(a, b *yamledit.KeyValue[string]) int {
			if a.Key == b.Key {
				return 0
			}
			if a.Key < b.Key {
				return -1
			}
			return 1
		}),
	))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
	// Output:
	// age: 10
	// job: engineer
	// name: foo # keep comment
}

func TestSortKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		action  yamledit.Action
		want    string
		wantErr bool
	}{
		{
			name: "sort alphabetically",
			yml: `name: foo
age: 10
job: engineer
`,
			action: yamledit.MapAction("$", yamledit.SortKey(ascKeyFunc)),
			want: `age: 10
job: engineer
name: foo
`,
		},
		{
			name: "already sorted",
			yml: `age: 10
job: engineer
name: foo
`,
			action: yamledit.MapAction("$", yamledit.SortKey(ascKeyFunc)),
			want: `age: 10
job: engineer
name: foo
`,
		},
		{
			name: "reverse sort",
			yml: `age: 10
job: engineer
name: foo
`,
			action: yamledit.MapAction("$", yamledit.SortKey(descKeyFunc)),
			want: `name: foo
job: engineer
age: 10
`,
		},
		{
			name: "single key",
			yml: `name: foo
`,
			action: yamledit.MapAction("$", yamledit.SortKey(ascKeyFunc)),
			want: `name: foo
`,
		},
		{
			name: "preserve comments",
			yml: `name: foo # comment1
age: 10 # comment2
`,
			action: yamledit.MapAction("$", yamledit.SortKey(ascKeyFunc)),
			want: `age: 10 # comment2
name: foo # comment1
`,
		},
		{
			name: "nested path",
			yml: `foo:
  c: 3
  a: 1
  b: 2
`,
			action: yamledit.MapAction("$.foo", yamledit.SortKey(ascKeyFunc)),
			want: `foo:
  a: 1
  b: 2
  c: 3
`,
		},
		{
			name:    "invalid yaml path",
			yml:     `name: foo`,
			action:  yamledit.MapAction("invalid[", yamledit.SortKey(ascKeyFunc)),
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
