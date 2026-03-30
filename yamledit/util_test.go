package yamledit_test

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-yaml/ast"
	"github.com/suzuki-shunsuke/go-yamledit/yamledit"
)

type errAction struct{}

func (a errAction) Run(_ ast.Node) error {
	return errors.New("action error")
}

func TestEditBytes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		yml     string
		actions []yamledit.Action
		want    string
		wantErr bool
	}{
		{
			name: "no actions",
			yml: `name: foo
age: 10
`,
			want: `name: foo
age: 10
`,
		},
		{
			name: "single action",
			yml: `name: foo
age: 10
`,
			actions: []yamledit.Action{
				yamledit.MapAction("$", yamledit.SetKey("name", "bar", nil)),
			},
			want: `name: bar
age: 10
`,
		},
		{
			name: "multiple actions",
			yml: `name: foo
age: 10
`,
			actions: []yamledit.Action{
				yamledit.MapAction("$", yamledit.SetKey("name", "bar", nil)),
				yamledit.MapAction("$", yamledit.RemoveKeys("age")),
			},
			want: `name: bar
`,
		},
		{
			name: "comments preserved",
			yml: `name: foo # keep this
age: 10
`,
			actions: []yamledit.Action{
				yamledit.MapAction("$", yamledit.SetKey("name", "bar", nil)),
			},
			want: `name: bar # keep this
age: 10
`,
		},
		{
			name:    "invalid YAML",
			yml:     `{invalid: [`,
			wantErr: true,
		},
		{
			name: "action error",
			yml: `name: foo
`,
			actions: []yamledit.Action{
				errAction{},
			},
			wantErr: true,
		},
		{
			name: "multi-document YAML",
			yml: `name: foo
---
name: bar
`,
			actions: []yamledit.Action{
				yamledit.MapAction("$", yamledit.SetKey("name", "updated", nil)),
			},
			want: `name: updated
---
name: updated
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := yamledit.EditBytes("example.yaml", []byte(tt.yml), tt.actions...)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestEditFile(t *testing.T) { //nolint:cyclop
	t.Parallel()
	tests := []struct {
		name      string
		yml       string
		actions   []yamledit.Action
		want      string
		wantBool  bool
		wantErr   bool
		noFile    bool
		checkFile bool // if true, verify file content unchanged on error
	}{
		{
			name: "single action modifies file",
			yml: `name: foo
age: 10
`,
			actions: []yamledit.Action{
				yamledit.MapAction("$", yamledit.SetKey("name", "bar", nil)),
			},
			want: `name: bar
age: 10
`,
			wantBool: true,
		},
		{
			name: "no change",
			yml: `name: foo
`,
			actions: []yamledit.Action{
				yamledit.MapAction("$", yamledit.SetKey("name", "foo", nil)),
			},
			want: `name: foo
`,
			wantBool: false,
		},
		{
			name:    "file not found",
			noFile:  true,
			wantErr: true,
		},
		{
			name: "action error",
			yml: `name: foo
`,
			actions: []yamledit.Action{
				errAction{},
			},
			wantErr:   true,
			checkFile: true,
		},
		{
			name: "multiple actions",
			yml: `name: foo
age: 10
`,
			actions: []yamledit.Action{
				yamledit.MapAction("$", yamledit.SetKey("name", "bar", nil)),
				yamledit.MapAction("$", yamledit.RemoveKeys("age")),
			},
			want: `name: bar
`,
			wantBool: true,
		},
		{
			name: "comments preserved",
			yml: `name: foo # keep this
age: 10
`,
			actions: []yamledit.Action{
				yamledit.MapAction("$", yamledit.SetKey("name", "bar", nil)),
			},
			want: `name: bar # keep this
age: 10
`,
			wantBool: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			filePath := filepath.Join(t.TempDir(), "test.yaml")
			if tt.noFile {
				// Use a path that doesn't exist
				filePath = filepath.Join(t.TempDir(), "nonexistent.yaml")
			} else {
				if err := os.WriteFile(filePath, []byte(tt.yml), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			got, err := yamledit.EditFile(filePath, tt.actions...)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.checkFile {
					b, readErr := os.ReadFile(filePath)
					if readErr != nil {
						t.Fatal(readErr)
					}
					if string(b) != tt.yml {
						t.Errorf("file was modified on error\ngot:\n%s\nwant:\n%s", string(b), tt.yml)
					}
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.wantBool {
				t.Errorf("got modified=%v, want %v", got, tt.wantBool)
			}
			b, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", string(b), tt.want)
			}
		})
	}
}

func ExampleNewBytes() {
	yml := `
- foo # comment
`

	s, err := yamledit.EditBytes("example.yaml", []byte(yml), yamledit.ListAction("$", yamledit.AddValuesToList(0, yamledit.NewBytes([]byte("hello # world")))))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
	// Output:
	// - hello # world
	// - foo # comment
}
