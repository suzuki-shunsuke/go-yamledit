package yamldoc_test

import (
	"cmp"
	"slices"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/suzuki-shunsuke/go-yamledit/yamldoc"
)

func TestDocument(t *testing.T) { //nolint:maintidx,cyclop
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		edit     func(t *testing.T, doc *yamldoc.Document) error
		expected string
	}{
		{
			name: "round trip",
			input: `# head
name: foo # line
# before id
id: 1
tags:
  # before a
  - a # ca
  # before b
  - b
nested:
  x:
    - k: v
      # c
      j: 2
flow: {a: 1, b: [1, 2]}
script: |
  echo foo
  echo bar
# foot
`,
			edit: func(_ *testing.T, _ *yamldoc.Document) error { return nil },
		},
		{
			name: "update a value",
			input: `id: 1 # id
name: foo # name
`,
			edit: func(_ *testing.T, doc *yamldoc.Document) error {
				return doc.Set("name", "bar")
			},
			expected: `id: 1 # id
name: bar # name
`,
		},
		{
			name: "delete a key",
			input: `id: 1
# name
name: foo
age: 10
`,
			edit: func(_ *testing.T, doc *yamldoc.Document) error {
				doc.Delete("name")
				return nil
			},
			expected: `id: 1
age: 10
`,
		},
		{
			name: "add a key",
			input: `id: 1
`,
			edit: func(_ *testing.T, doc *yamldoc.Document) error {
				return doc.Set("name", "foo")
			},
			expected: `id: 1
name: foo
`,
		},
		{
			name: "insert a pair at a position",
			input: `name: foo
age: 10
`,
			edit: func(_ *testing.T, doc *yamldoc.Document) error {
				m := doc.Contents.(*yamldoc.Map)
				p, err := yamldoc.NewPair("id", 123)
				if err != nil {
					return err
				}
				m.Items = slices.Insert(m.Items, m.IndexOf("name")+1, p)
				return nil
			},
			expected: `name: foo
id: 123
age: 10
`,
		},
		{
			name: "sort keys",
			input: `c: 3 # c
# before b
b: 2
a:
  z: 1
  y: 2
`,
			edit: func(_ *testing.T, doc *yamldoc.Document) error {
				m := doc.Contents.(*yamldoc.Map)
				slices.SortFunc(m.Items, func(a, b *yamldoc.Pair) int {
					return cmp.Compare(a.KeyString(), b.KeyString())
				})
				return nil
			},
			expected: `a:
  z: 1
  y: 2
# before b
b: 2
c: 3 # c
`,
		},
		{
			name: "rename a key",
			input: `name: foo # comment
age: 10
`,
			edit: func(_ *testing.T, doc *yamldoc.Document) error {
				doc.Contents.(*yamldoc.Map).Pair("name").Key.Value = "first_name"
				return nil
			},
			expected: `first_name: foo # comment
age: 10
`,
		},
		{
			name: "edit a sequence",
			input: `tags:
  - a
  - b # b
  - c
`,
			edit: func(_ *testing.T, doc *yamldoc.Document) error {
				tags := doc.Get("tags").(*yamldoc.Seq)
				tags.Items = slices.Delete(tags.Items, 0, 1)
				n, err := yamldoc.NewNode("x")
				if err != nil {
					return err
				}
				n.SetComment(" x")
				n.SetCommentBefore(" before x")
				tags.Items = slices.Insert(tags.Items, 1, n)
				tags.Items[0].SetComment(" tag b")
				return tags.Add(map[string]any{"name": "d", "value": 4})
			},
			expected: `tags:
  - b # tag b
  # before x
  - x # x
  - c
  - name: d
    value: 4
`,
		},
		{
			name: "comments",
			input: `a: 1
b: 2
`,
			edit: func(_ *testing.T, doc *yamldoc.Document) error {
				m := doc.Contents.(*yamldoc.Map)
				m.Pair("b").SetCommentBefore(" before b\n second line")
				m.Get("a").SetComment(" a")
				return nil
			},
			expected: `a: 1 # a
# before b
# second line
b: 2
`,
		},
		{
			name: "set nested map with the document's indentation",
			input: `jobs:
    test:
        steps:
        - uses: actions/checkout@v6
`,
			edit: func(_ *testing.T, doc *yamldoc.Document) error {
				return doc.SetIn([]any{"jobs", "test", "steps", 0, "with"}, yaml.MapSlice{
					{Key: "persist-credentials", Value: false},
					{Key: "paths", Value: []string{"a", "b"}},
				})
			},
			expected: `jobs:
    test:
        steps:
        - uses: actions/checkout@v6
          with:
              persist-credentials: false
              paths:
              - a
              - b
`,
		},
		{
			name: "SetIn creates intermediate maps",
			input: `a: 1
`,
			edit: func(_ *testing.T, doc *yamldoc.Document) error {
				return doc.SetIn([]any{"b", "c", "d"}, 1)
			},
			expected: `a: 1
b:
  c:
    d: 1
`,
		},
		{
			name: "move a subtree to another depth",
			input: `a:
  b:
    c: 1
    d: |
      foo
      bar
`,
			edit: func(_ *testing.T, doc *yamldoc.Document) error {
				root := doc.Contents.(*yamldoc.Map)
				b := root.GetIn("a", "b")
				root.DeleteIn("a", "b")
				return root.Set("b", b)
			},
			expected: `a: {}
b:
  c: 1
  d: |
    foo
    bar
`,
		},
		{
			name: "add to a flow map",
			input: `a: {x: 1}
`,
			edit: func(_ *testing.T, doc *yamldoc.Document) error {
				return doc.GetIn("a").(*yamldoc.Map).Set("w", map[string]any{"z": []int{1, 2}})
			},
			expected: `a: {x: 1, w: {z: [1, 2]}}
`,
		},
		{
			name: "set a multi-line string",
			input: `a:
  b: 1
`,
			edit: func(_ *testing.T, doc *yamldoc.Document) error {
				return doc.SetIn([]any{"a", "b"}, "foo\nbar\n")
			},
			expected: `a:
  b: |
    foo
    bar
`,
		},
		{
			name:  "empty document",
			input: ``,
			edit: func(_ *testing.T, doc *yamldoc.Document) error {
				return doc.Set("a", []string{"x"})
			},
			expected: `a:
  - x
`,
		},
		{
			name: "anchor",
			input: `a: &a
  x: 1
b: *a
`,
			edit: func(t *testing.T, doc *yamldoc.Document) error {
				t.Helper()
				a := doc.Get("a").(*yamldoc.Map)
				if a.Anchor() != "a" {
					t.Errorf("anchor: got %q", a.Anchor())
				}
				if n := doc.Get("b").(*yamldoc.Alias).Name(); n != "a" {
					t.Errorf("alias: got %q", n)
				}
				return a.Set("z", 2)
			},
			expected: `a: &a
  x: 1
  z: 2
b: *a
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			doc, err := yamldoc.Parse([]byte(tt.input))
			if err != nil {
				t.Fatal(err)
			}
			if err := tt.edit(t, doc); err != nil {
				t.Fatal(err)
			}
			s, err := doc.ToString()
			if err != nil {
				t.Fatal(err)
			}
			expected := tt.expected
			if expected == "" {
				expected = tt.input
			}
			if s != expected {
				t.Errorf("got:\n%s\nexpected:\n%s", s, expected)
			}
		})
	}
}

// TestSetKeyInSteps is the scenario of https://github.com/suzuki-shunsuke/go-yamledit tasks/bug-set-key.md
func TestSetKeyInSteps(t *testing.T) {
	t.Parallel()
	input := `jobs:
  test:
    runs-on: ubuntu-24.04 # comment
    steps:
      - uses: actions/checkout@v6
        # with:
        #   persist-credentials: false
---
jobs:
  test:
    runs-on: ubuntu-24.04 # comment
    steps:
      - uses: actions/checkout@v6
        with:
          persist-credentials: false
---
jobs:
  test:
    runs-on: ubuntu-24.04 # comment
    steps:
      - uses: actions/checkout@v6
        with:
          token: ${{github.token}}
          persist-credentials: true
`
	expected := `jobs:
  test:
    runs-on: ubuntu-24.04 # comment
    steps:
      - uses: actions/checkout@v6
        # with:
        #   persist-credentials: false
        with:
          persist-credentials: false
---
jobs:
  test:
    runs-on: ubuntu-24.04 # comment
    steps:
      - uses: actions/checkout@v6
        with:
          persist-credentials: false
---
jobs:
  test:
    runs-on: ubuntu-24.04 # comment
    steps:
      - uses: actions/checkout@v6
        with:
          token: ${{github.token}}
          persist-credentials: false
`
	f, err := yamldoc.ParseAll([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	for _, doc := range f.Docs {
		for _, step := range workflowSteps(doc) {
			if err := step.SetIn([]any{"with", "persist-credentials"}, false); err != nil {
				t.Fatal(err)
			}
		}
	}
	s, err := f.ToString()
	if err != nil {
		t.Fatal(err)
	}
	if s != expected {
		t.Errorf("got:\n%s\nexpected:\n%s", s, expected)
	}
}

func TestGetAs(t *testing.T) {
	t.Parallel()
	doc, err := yamldoc.Parse([]byte(`jobs:
  test:
    runs-on: ubuntu-24.04
    steps:
      - uses: actions/checkout@v6
`))
	if err != nil {
		t.Fatal(err)
	}
	s, ok, err := yamldoc.GetAs[string](doc.Contents, "jobs", "test", "steps", 0, "uses")
	if err != nil {
		t.Fatal(err)
	}
	if !ok || s != "actions/checkout@v6" {
		t.Errorf("got %q, %v", s, ok)
	}
	type Step struct {
		Uses string `yaml:"uses"`
	}
	steps, _, err := yamldoc.GetAs[[]Step](doc.Contents, "jobs", "test", "steps")
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 1 || steps[0].Uses != "actions/checkout@v6" {
		t.Errorf("got %+v", steps)
	}
	if _, ok, _ := yamldoc.GetAs[string](doc.Contents, "jobs", "foo"); ok {
		t.Error("expected not found")
	}
}

func TestCommentOnlyDocument(t *testing.T) {
	t.Parallel()
	doc, err := yamldoc.Parse([]byte("# license\n"))
	if err != nil {
		t.Fatal(err)
	}
	s, err := doc.ToString()
	if err != nil {
		t.Fatal(err)
	}
	if s != "# license\n" {
		t.Errorf("got %q", s)
	}
	if err := doc.Set("a", 1); err != nil {
		t.Fatal(err)
	}
	s, err = doc.ToString()
	if err != nil {
		t.Fatal(err)
	}
	if s != "# license\na: 1\n" {
		t.Errorf("got %q", s)
	}
}

func TestEdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		edit     func(doc *yamldoc.Document) error
		expected string
	}{
		{
			name:  "an empty flow map becomes block style",
			input: "with: {}\n",
			edit: func(doc *yamldoc.Document) error {
				return doc.SetIn([]any{"with", "persist-credentials"}, false)
			},
			expected: "with:\n  persist-credentials: false\n",
		},
		{
			name:  "an empty flow sequence becomes block style",
			input: "a:\n  tags: []\n",
			edit: func(doc *yamldoc.Document) error {
				return doc.GetIn("a", "tags").(*yamldoc.Seq).Add("x")
			},
			expected: "a:\n  tags:\n    - x\n",
		},
		{
			name:  "a multi-line string in a flow sequence is quoted",
			input: "a: [x]\n",
			edit: func(doc *yamldoc.Document) error {
				return doc.Get("a").(*yamldoc.Seq).Add("foo\nbar\n")
			},
			expected: "a: [x, \"foo\\nbar\\n\"]\n",
		},
		{
			name:  "a block scalar at the end of the file without a line break",
			input: "a: |\n  foo",
			edit: func(doc *yamldoc.Document) error {
				return doc.Set("b", 1)
			},
			expected: "a: |-\n  foo\nb: 1\n",
		},
		{
			name:  "no indentation of sequences",
			input: "a:\n- x\n",
			edit: func(doc *yamldoc.Document) error {
				return doc.Set("b", []string{"z"})
			},
			expected: "a:\n- x\nb:\n- z\n",
		},
		{
			name:  "the comment after the last pair stays with the pair",
			input: "a: 1\n# after a\n",
			edit: func(doc *yamldoc.Document) error {
				return doc.Set("b", 2)
			},
			expected: "a: 1\n# after a\nb: 2\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			doc, err := yamldoc.Parse([]byte(tt.input))
			if err != nil {
				t.Fatal(err)
			}
			if err := tt.edit(doc); err != nil {
				t.Fatal(err)
			}
			s, err := doc.ToString()
			if err != nil {
				t.Fatal(err)
			}
			if s != tt.expected {
				t.Errorf("got:\n%s\nexpected:\n%s", s, tt.expected)
			}
		})
	}
}

func TestToStringTwice(t *testing.T) {
	t.Parallel()
	doc, err := yamldoc.Parse([]byte("a:\n  b: 1\n"))
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.SetIn([]any{"a", "c"}, map[string]any{"d": "foo\nbar\n"}); err != nil {
		t.Fatal(err)
	}
	s1, err := doc.ToString()
	if err != nil {
		t.Fatal(err)
	}
	s2, err := doc.ToString()
	if err != nil {
		t.Fatal(err)
	}
	if s1 != s2 {
		t.Errorf("ToString isn't idempotent:\n%s\n---\n%s", s1, s2)
	}
	// Move the nested map to the root after the first ToString.
	c := doc.GetIn("a", "c")
	doc.DeleteIn("a", "c")
	if err := doc.Set("c", c); err != nil {
		t.Fatal(err)
	}
	s3, err := doc.ToString()
	if err != nil {
		t.Fatal(err)
	}
	expected := "a:\n  b: 1\nc:\n  d: |\n    foo\n    bar\n"
	if s3 != expected {
		t.Errorf("got:\n%s\nexpected:\n%s", s3, expected)
	}
}

// workflowSteps returns steps of all jobs in a GitHub Actions workflow.
func workflowSteps(doc *yamldoc.Document) []*yamldoc.Map {
	jobs, ok := doc.Get("jobs").(*yamldoc.Map)
	if !ok {
		return nil
	}
	var ret []*yamldoc.Map
	for _, job := range jobs.Items {
		steps, ok := job.Value.(*yamldoc.Map).Get("steps").(*yamldoc.Seq)
		if !ok {
			continue
		}
		for _, step := range steps.Items {
			if m, ok := step.(*yamldoc.Map); ok {
				ret = append(ret, m)
			}
		}
	}
	return ret
}
