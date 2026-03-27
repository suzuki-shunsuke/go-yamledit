package mag_test

import (
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleEditMap() {
	yml := `
name: foo # keep comment
age: 10 # keep comment 2
type: yoo # keep comment 3
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	act := mag.Map(
		"$",
		mag.EditMapAction[string, any](
			// Change the value of the "name" key to "new name"
			func(m *mag.MapValue[string, any]) error {
				kv, ok := m.Map["name"]
				if !ok {
					return nil
				}
				return mag.SetValueToMappingValue(kv.Node, "new name", false)
			},
		),
		mag.EditMapAction[string, any](
			// If the given key does not exist, do nothing
			func(m *mag.MapValue[string, any]) error {
				kv, ok := m.Map["password"]
				if !ok {
					return nil
				}
				return mag.SetValueToMappingValue(kv.Node, "***", false)
			},
		),
		mag.EditMapAction[string, any](
			// Rename the "age" key to "age-2"
			func(m *mag.MapValue[string, any]) error {
				kv, ok := m.Map["age"]
				if !ok {
					return nil
				}
				return mag.RenameKeyOfMappingValueNode(kv.Node, "age-2")
			},
		),
		mag.EditMapAction[string, any](
			// Change both key and value
			// key: type => type-2
			// value yoo => yoo-2
			func(m *mag.MapValue[string, any]) error {
				kv, ok := m.Map["type"]
				if !ok {
					return nil
				}
				if err := mag.RenameKeyOfMappingValueNode(kv.Node, "type-2"); err != nil {
					return err
				}
				return mag.SetValueToMappingValue(kv.Node, "yoo-2", false)
			},
		),
	)
	if err := act.Run(file.Docs[0].Body); err != nil {
		log.Fatal(err)
	}
	fmt.Println(file.String())
	// Output:
	// name: new name # keep comment
	// age-2: 10 # keep comment 2
	// type-2: yoo-2 # keep comment 3
}

func TestEditMapAction_Run(t *testing.T) { //nolint:cyclop
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		action  mag.Action
		want    string
		wantErr bool
	}{
		{
			name: "set value",
			yml: `name: foo
age: 10
`,
			action: mag.Map("$", mag.EditMapAction[string, any](
				func(m *mag.MapValue[string, any]) error {
					kv, ok := m.Map["name"]
					if !ok {
						return nil
					}
					return mag.SetValueToMappingValue(kv.Node, "bar", false)
				},
			)),
			want: `name: bar
age: 10
`,
		},
		{
			name: "rename key",
			yml: `name: foo
age: 10
`,
			action: mag.Map("$", mag.EditMapAction[string, any](
				func(m *mag.MapValue[string, any]) error {
					kv, ok := m.Map["name"]
					if !ok {
						return nil
					}
					return mag.RenameKeyOfMappingValueNode(kv.Node, "first_name")
				},
			)),
			want: `first_name: foo
age: 10
`,
		},
		{
			name: "key not found",
			yml: `name: foo
`,
			action: mag.Map("$", mag.EditMapAction[string, any](
				func(m *mag.MapValue[string, any]) error {
					_, ok := m.Map["missing"]
					if !ok {
						return nil
					}
					return nil
				},
			)),
			want: `name: foo
`,
		},
		{
			name: "preserve comment",
			yml: `name: foo # important
`,
			action: mag.Map("$", mag.EditMapAction[string, any](
				func(m *mag.MapValue[string, any]) error {
					kv := m.Map["name"]
					return mag.SetValueToMappingValue(kv.Node, "bar", false)
				},
			)),
			want: `name: bar # important
`,
		},
		{
			name: "edit func returns error",
			yml: `name: foo
`,
			action: mag.Map("$", mag.EditMapAction[string, any](
				func(_ *mag.MapValue[string, any]) error {
					return errors.New("edit error")
				},
			)),
			wantErr: true,
		},
		{
			name: "multiple changes",
			yml: `name: foo
age: 10
`,
			action: mag.Map("$", mag.EditMapAction[string, any](
				func(m *mag.MapValue[string, any]) error {
					if err := mag.SetValueToMappingValue(m.Map["name"].Node, "bar", false); err != nil {
						return err
					}
					return mag.SetValueToMappingValue(m.Map["age"].Node, 20, false)
				},
			)),
			want: `name: bar
age: 20
`,
		},
		{
			name: "no changes",
			yml: `name: foo
`,
			action: mag.Map("$", mag.EditMapAction[string, any](
				func(_ *mag.MapValue[string, any]) error {
					return nil
				},
			)),
			want: `name: foo
`,
		},
		{
			name: "nested path",
			yml: `foo:
  bar: 1
  baz: 2
`,
			action: mag.Map("$.foo", mag.EditMapAction[string, any](
				func(m *mag.MapValue[string, any]) error {
					kv := m.Map["bar"]
					return mag.SetValueToMappingValue(kv.Node, 99, false)
				},
			)),
			want: `foo:
  bar: 99
  baz: 2
`,
		},
		{
			name: "invalid yaml path",
			yml:  `name: foo`,
			action: mag.Map("invalid[", mag.EditMapAction[string, any](
				func(_ *mag.MapValue[string, any]) error {
					return nil
				},
			)),
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
