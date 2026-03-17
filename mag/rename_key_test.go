package mag_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleRenameKeyAction_Run() {
	yml := `
name: foo # keep comment
age: 10 # keep comment 2
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	actions := []mag.Action{
		&mag.RenameKeyAction{
			YAMLPath: "$",
			OldKey:   "name",
			NewKey:   "id",
		},
		&mag.RenameKeyAction{
			YAMLPath: "$",
			OldKey:   "password", // unknown key
			NewKey:   "passwd",
		},
	}
	for _, act := range actions {
		if err := act.Run(file.Docs[0].Body); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(file.String())
	// Output:
	// id: foo # keep comment
	// age: 10 # keep comment 2
}

func TestRenameKeyAction_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		action  mag.RenameKeyAction
		want    string
		wantErr bool
	}{
		{
			name: "rename root key",
			yml:  "name: foo\nage: 10\n",
			action: mag.RenameKeyAction{
				YAMLPath: "$",
				OldKey:   "name",
				NewKey:   "id",
			},
			want: "id: foo\nage: 10\n",
		},
		{
			name: "key not found",
			yml:  "name: foo\n",
			action: mag.RenameKeyAction{
				YAMLPath: "$",
				OldKey:   "missing",
				NewKey:   "found",
			},
			want: "name: foo\n",
		},
		{
			name: "nested path",
			yml:  "foo:\n  bar: 1\n  baz: 2\n",
			action: mag.RenameKeyAction{
				YAMLPath: "$.foo",
				OldKey:   "bar",
				NewKey:   "qux",
			},
			// The YAML library resets indentation on the renamed key node.
			want: "foo:\nqux: 1\n  baz: 2\n",
		},
		{
			name: "sequence of mappings",
			yml:  "items:\n- name: a\n  val: 1\n- name: b\n  val: 2\n",
			action: mag.RenameKeyAction{
				YAMLPath: "$.items",
				OldKey:   "name",
				NewKey:   "id",
			},
			want: "items:\n- id: a\n    val: 1\n- id: b\n    val: 2\n",
		},
		{
			name: "invalid yaml path",
			yml:  "name: foo\n",
			action: mag.RenameKeyAction{
				YAMLPath: "invalid[",
				OldKey:   "name",
				NewKey:   "id",
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
