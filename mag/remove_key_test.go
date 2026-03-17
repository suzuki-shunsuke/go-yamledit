package mag_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/mag-go-sdk/mag"
)

func ExampleRemoveKeyAction_Run() {
	yml := `
name: foo
age: 10 # keep comment
`

	file, err := parser.ParseBytes([]byte(yml), parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	actions := []mag.Action{
		&mag.RemoveKeyAction{
			YAMLPath: "$",
			Key:      "name",
		},
		&mag.RemoveKeyAction{
			YAMLPath: "$",
			Key:      "id", // unknown key
		},
	}
	for _, act := range actions {
		if err := act.Run(file.Docs[0].Body); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(file.String())
	// Output:
	// age: 10 # keep comment
}

func TestRemoveKeyAction_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		action  mag.RemoveKeyAction
		want    string
		wantErr bool
	}{
		{
			name: "remove root key",
			yml:  "name: foo\nage: 10\n",
			action: mag.RemoveKeyAction{
				YAMLPath: "$",
				Key:      "name",
			},
			want: "age: 10\n",
		},
		{
			name: "key not found",
			yml:  "name: foo\n",
			action: mag.RemoveKeyAction{
				YAMLPath: "$",
				Key:      "missing",
			},
			want: "name: foo\n",
		},
		{
			name: "nested path",
			yml:  "foo:\n  bar: 1\n  baz: 2\n",
			action: mag.RemoveKeyAction{
				YAMLPath: "$.foo",
				Key:      "bar",
			},
			want: "foo:\n  baz: 2\n",
		},
		{
			name: "sequence of mappings",
			yml:  "items:\n- name: a\n  val: 1\n- name: b\n  val: 2\n",
			action: mag.RemoveKeyAction{
				YAMLPath: "$.items",
				Key:      "name",
			},
			want: "items:\n- val: 1\n- val: 2\n",
		},
		{
			name: "invalid yaml path",
			yml:  "name: foo\n",
			action: mag.RemoveKeyAction{
				YAMLPath: "invalid[",
				Key:      "name",
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
