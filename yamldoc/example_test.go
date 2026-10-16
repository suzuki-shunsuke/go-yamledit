package yamldoc_test

import (
	"cmp"
	"fmt"
	"log"
	"slices"

	"github.com/suzuki-shunsuke/go-yamledit/yamldoc"
)

func Example() {
	doc, err := yamldoc.Parse([]byte(`# user
name: foo # name
id: 1
tags:
  - b
  - a
`))
	if err != nil {
		log.Fatal(err)
	}
	// Update a value keeping the comment.
	if err := doc.Set("name", "bar"); err != nil {
		log.Fatal(err)
	}
	// Delete a key.
	doc.Delete("id")
	// Add a key.
	if err := doc.Set("age", 20); err != nil {
		log.Fatal(err)
	}

	m, ok := doc.Contents.(*yamldoc.Map)
	if !ok {
		log.Fatal("the root must be a map")
	}
	// Insert a pair at a specific position.
	p, err := yamldoc.NewPair("id", 123)
	if err != nil {
		log.Fatal(err)
	}
	m.Items = slices.Insert(m.Items, m.IndexOf("name")+1, p)
	// Rename a key.
	m.Pair("age").Key.Value = "years"

	// Sort a sequence and add a comment.
	tags, ok := doc.Get("tags").(*yamldoc.Seq)
	if !ok {
		log.Fatal("tags must be a sequence")
	}
	slices.SortFunc(tags.Items, func(a, b yamldoc.Node) int {
		return cmp.Compare(scalarString(a), scalarString(b))
	})
	tags.Items[1].SetComment(" tag b")

	s, err := doc.ToString()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(s)
	// Output:
	// # user
	// name: bar # name
	// id: 123
	// tags:
	//   - a
	//   - b # tag b
	// years: 20
}

func scalarString(n yamldoc.Node) string {
	if s, ok := n.(*yamldoc.Scalar); ok {
		return s.String()
	}
	return ""
}

func ExampleDocument_SetIn() {
	doc, err := yamldoc.Parse([]byte(`jobs:
  test:
    steps:
      - uses: actions/checkout@v6
`))
	if err != nil {
		log.Fatal(err)
	}
	if err := doc.SetIn([]any{"jobs", "test", "steps", 0, "with", "persist-credentials"}, false); err != nil {
		log.Fatal(err)
	}
	s, err := doc.ToString()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(s)
	// Output:
	// jobs:
	//   test:
	//     steps:
	//       - uses: actions/checkout@v6
	//         with:
	//           persist-credentials: false
}

func ExampleGetAs() {
	doc, err := yamldoc.Parse([]byte(`jobs:
  test:
    runs-on: ubuntu-24.04
`))
	if err != nil {
		log.Fatal(err)
	}
	runsOn, ok, err := yamldoc.GetAs[string](doc.Contents, "jobs", "test", "runs-on")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(runsOn, ok)
	// Output:
	// ubuntu-24.04 true
}
