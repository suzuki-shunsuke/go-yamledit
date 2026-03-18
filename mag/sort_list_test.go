package mag_test

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleSortListAction_Run() {
	yml := `
- foo # comment
- bar # comment 2
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	act := &mag.SortListAction[string]{
		YAMLPath: "$",
		Sort: func(a, b *mag.Item[string]) int {
			return strings.Compare(a.Value, b.Value)
		},
	}
	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// - bar # comment 2
	// - foo # comment
}

func TestSortListAction_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		action  mag.SortListAction[string]
		want    string
		wantErr bool
	}{
		{
			name: "sort strings alphabetically",
			yml: `- cherry
- apple
- banana
`,
			action: mag.SortListAction[string]{
				YAMLPath: "$",
				Sort: func(a, b *mag.Item[string]) int {
					return strings.Compare(a.Value, b.Value)
				},
			},
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
			action: mag.SortListAction[string]{
				YAMLPath: "$",
				Sort: func(a, b *mag.Item[string]) int {
					return strings.Compare(a.Value, b.Value)
				},
			},
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
			action: mag.SortListAction[string]{
				YAMLPath: "$",
				Sort: func(a, b *mag.Item[string]) int {
					return strings.Compare(b.Value, a.Value)
				},
			},
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
			action: mag.SortListAction[string]{
				YAMLPath: "$",
				Sort: func(a, b *mag.Item[string]) int {
					return strings.Compare(a.Value, b.Value)
				},
			},
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
			action: mag.SortListAction[string]{
				YAMLPath: "$.foo.items",
				Sort: func(a, b *mag.Item[string]) int {
					return strings.Compare(a.Value, b.Value)
				},
			},
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
			action: mag.SortListAction[string]{
				YAMLPath: "$.items[*]",
				Sort: func(a, b *mag.Item[string]) int {
					return strings.Compare(a.Value, b.Value)
				},
			},
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
			action: mag.SortListAction[string]{
				YAMLPath: "$",
				Sort: func(a, b *mag.Item[string]) int {
					return strings.Compare(a.Value, b.Value)
				},
			},
			want: `- only
`,
		},
		{
			name: "invalid yaml path",
			yml: `- a
`,
			action: mag.SortListAction[string]{
				YAMLPath: "invalid[",
				Sort: func(_, _ *mag.Item[string]) int {
					return 0
				},
			},
			wantErr: true,
		},
		{
			name: "Sort is nil",
			yml: `- a
`,
			action: mag.SortListAction[string]{
				YAMLPath: "$",
			},
			wantErr: true,
		},
		{
			name: "YAMLPath is empty",
			yml: `- a
`,
			action: mag.SortListAction[string]{
				Sort: func(_, _ *mag.Item[string]) int {
					return 0
				},
			},
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
		action  mag.SortListAction[uint64]
		want    string
		wantErr bool
	}{
		{
			name: "sort integers",
			yml: `- 3
- 1
- 2
`,
			action: mag.SortListAction[uint64]{
				YAMLPath: "$",
				Sort: func(a, b *mag.Item[uint64]) int {
					aInt := a.Value
					bInt := b.Value
					if aInt < bInt {
						return -1
					}
					if aInt > bInt {
						return 1
					}
					return 0
				},
			},
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
