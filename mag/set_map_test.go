package mag_test

import (
	"fmt"
	"log"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
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
	act := mag.Map(
		"$",
		// Edit name to "ryan"
		mag.SetKey("name", "ryan", nil),
		mag.SetKey("gender", "male", &mag.SetKeyOption{
			InsertLocations: []*mag.InsertLocation{
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
