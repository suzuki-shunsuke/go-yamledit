package mag_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleEditMapValueAction_Run() {
	yml := `
name: foo # keep comment
age: 10 # keep comment 2
type: yoo # keep comment 3
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	actions := []mag.Action{
		&mag.EditMapValueAction{
			// Change the value of the "name" key to "new name"
			YAMLPath: "$",
			Matcher:  mag.NewKeyMVMatcher("name"),
			Editor:   mag.NewStaticMappingValueEditor(mag.Noop, "new name"),
		},
		&mag.EditMapValueAction{
			// If the given key does not exist, do nothing
			YAMLPath: "$",
			Matcher:  mag.NewKeyMVMatcher("password"), // unknown key
			Editor:   mag.NewStaticMappingValueEditor(mag.Noop, "***"),
		},
		&mag.EditMapValueAction{
			// Rename the "age" key to "age-2"
			YAMLPath: "$",
			Matcher:  mag.NewKeyMVMatcher("age"),
			Editor:   mag.NewStaticMappingValueEditor("age-2", mag.Noop),
		},
		&mag.EditMapValueAction{
			// Change both key and value
			// key: type => type-2
			// value yoo => yoo-2
			YAMLPath: "$",
			Matcher:  mag.NewKeyMVMatcher("type"),
			Editor:   mag.NewStaticMappingValueEditor("type-2", "yoo-2"),
		},
	}
	for _, act := range actions {
		if err := act.Run(file.Docs[0].Body); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(file.String())
	// Output:
	// name: new name # keep comment
	// age-2: 10 # keep comment 2
	// type-2: yoo-2 # keep comment 3
}

func TestEditMapValueAction_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		action  mag.EditMapValueAction
		want    string
		wantErr bool
	}{
		{
			name: "update root value",
			yml:  "name: foo\nage: 10\n",
			action: mag.EditMapValueAction{
				YAMLPath: "$",
				Matcher:  mag.NewKeyMVMatcher("name"),
				Editor:   mag.NewStaticMappingValueEditor(mag.Noop, "bar"),
			},
			want: "name: bar\nage: 10\n",
		},
		{
			name: "key not found",
			yml:  "name: foo\n",
			action: mag.EditMapValueAction{
				YAMLPath: "$",
				Matcher:  mag.NewKeyMVMatcher("missing"),
				Editor:   mag.NewStaticMappingValueEditor(mag.Noop, "val"),
			},
			want: "name: foo\n",
		},
		{
			name: "nested path",
			yml:  "foo:\n  bar: 1\n  baz: 2\n",
			action: mag.EditMapValueAction{
				YAMLPath: "$.foo",
				Matcher:  mag.NewKeyMVMatcher("bar"),
				Editor:   mag.NewStaticMappingValueEditor(mag.Noop, 99),
			},
			want: "foo:\n  bar: 99\n  baz: 2\n",
		},
		{
			name: "sequence of mappings",
			yml:  "items:\n- name: a\n  val: 1\n- name: b\n  val: 2\n",
			action: mag.EditMapValueAction{
				YAMLPath: "$.items",
				Matcher:  mag.NewKeyMVMatcher("val"),
				Editor:   mag.NewStaticMappingValueEditor(mag.Noop, 100),
			},
			want: "items:\n- name: a\n  val: 100\n- name: b\n  val: 100\n",
		},
		{
			name: "invalid yaml path",
			yml:  "name: foo\n",
			action: mag.EditMapValueAction{
				YAMLPath: "invalid[",
				Matcher:  mag.NewKeyMVMatcher("name"),
				Editor:   mag.NewStaticMappingValueEditor(mag.Noop, "bar"),
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
