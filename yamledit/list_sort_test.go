package yamledit_test

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
)

func ExampleSortList() {
	yml := `
- foo # comment
- bar # comment 2
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	act := yamledit.ListAction(
		"$",
		yamledit.SortList[string](func(a, b *yamledit.Node[string]) int {
			return strings.Compare(a.Value, b.Value)
		}),
	)
	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// - bar # comment 2
	// - foo # comment
}

func TestSortList(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		action  yamledit.Action
		want    string
		wantErr bool
	}{
		{
			name: "sort strings alphabetically",
			yml: `- cherry
- apple
- banana
`,
			action: yamledit.ListAction("$", yamledit.SortList[string](func(a, b *yamledit.Node[string]) int {
				return strings.Compare(a.Value, b.Value)
			})),
			want: `- apple
- banana
- cherry
`,
		},
		{
			name: "already sorted",
			yml: `- a
- b
- c
`,
			action: yamledit.ListAction("$", yamledit.SortList[string](func(a, b *yamledit.Node[string]) int {
				return strings.Compare(a.Value, b.Value)
			})),
			want: `- a
- b
- c
`,
		},
		{
			name: "reverse sort",
			yml: `- a
- b
- c
`,
			action: yamledit.ListAction("$", yamledit.SortList[string](func(a, b *yamledit.Node[string]) int {
				return strings.Compare(b.Value, a.Value)
			})),
			want: `- c
- b
- a
`,
		},
		{
			name: "with comment preservation",
			yml: `- cherry # third
- apple # first
- banana # second
`,
			action: yamledit.ListAction("$", yamledit.SortList[string](func(a, b *yamledit.Node[string]) int {
				return strings.Compare(a.Value, b.Value)
			})),
			want: `- apple # first
- banana # second
- cherry # third
`,
		},
		{
			name: "nested path",
			yml: `foo:
  items:
  - c
  - a
  - b
`,
			action: yamledit.ListAction("$.foo.items", yamledit.SortList[string](func(a, b *yamledit.Node[string]) int {
				return strings.Compare(a.Value, b.Value)
			})),
			want: `foo:
  items:
  - a
  - b
  - c
`,
		},
		{
			name: "sequence of sequences",
			yml: `items:
- - c
  - a
  - b
- - z
  - x
  - y
`,
			action: yamledit.ListAction("$.items[*]", yamledit.SortList[string](func(a, b *yamledit.Node[string]) int {
				return strings.Compare(a.Value, b.Value)
			})),
			want: `items:
- - a
  - b
  - c
- - x
  - y
  - z
`,
		},
		{
			name: "single element",
			yml: `- only
`,
			action: yamledit.ListAction("$", yamledit.SortList[string](func(a, b *yamledit.Node[string]) int {
				return strings.Compare(a.Value, b.Value)
			})),
			want: `- only
`,
		},
		{
			name: "invalid yaml path",
			yml: `- a
`,
			action: yamledit.ListAction("invalid[", yamledit.SortList[string](func(a, b *yamledit.Node[string]) int {
				return strings.Compare(a.Value, b.Value)
			})),
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

func TestSortListAction_Run_uint64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		action  yamledit.Action
		want    string
		wantErr bool
	}{
		{
			name: "sort integers",
			yml: `- 3
- 1
- 2
`,
			action: yamledit.ListAction("$", yamledit.SortList[uint64](func(a, b *yamledit.Node[uint64]) int {
				aInt := a.Value
				bInt := b.Value
				if aInt < bInt {
					return -1
				}
				if aInt > bInt {
					return 1
				}
				return 0
			})),
			want: `- 1
- 2
- 3
`,
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
