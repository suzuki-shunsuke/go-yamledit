package mag_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ascKeyFunc(a, b *mag.KeyValue) int {
	ak, bk := a.Key.(string), b.Key.(string)
	if ak < bk {
		return -1
	}
	if ak > bk {
		return 1
	}
	return 0
}

func descKeyFunc(a, b *mag.KeyValue) int {
	return -ascKeyFunc(a, b)
}

func ExampleSortKey() {
	yml := `
name: foo # keep comment
age: 10
job: engineer
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	act := mag.Map(
		"$",
		mag.SortKey(func(a, b *mag.KeyValue) int {
			if a.Key == b.Key {
				return 0
			}
			if a.Key.(string) < b.Key.(string) {
				return -1
			}
			return 1
		}),
	)

	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
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
		action  mag.Action
		want    string
		wantErr bool
	}{
		{
			name: "sort alphabetically",
			yml: `name: foo
age: 10
job: engineer
`,
			action: mag.Map("$", mag.SortKey(ascKeyFunc)),
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
			action: mag.Map("$", mag.SortKey(ascKeyFunc)),
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
			action: mag.Map("$", mag.SortKey(descKeyFunc)),
			want: `name: foo
job: engineer
age: 10
`,
		},
		{
			name: "single key",
			yml: `name: foo
`,
			action: mag.Map("$", mag.SortKey(ascKeyFunc)),
			want: `name: foo
`,
		},
		{
			name: "preserve comments",
			yml: `name: foo # comment1
age: 10 # comment2
`,
			action: mag.Map("$", mag.SortKey(ascKeyFunc)),
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
			action: mag.Map("$.foo", mag.SortKey(ascKeyFunc)),
			want: `foo:
  a: 1
  b: 2
  c: 3
`,
		},
		{
			name:    "invalid yaml path",
			yml:     `name: foo`,
			action:  mag.Map("invalid[", mag.SortKey(ascKeyFunc)),
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
