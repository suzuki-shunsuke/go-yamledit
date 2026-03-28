package yamledit

import (
	"testing"
)

func Test_unifyInt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    any
		want     any
		wantBool bool
	}{
		{
			name:     "int64",
			input:    int64(42),
			want:     "42",
			wantBool: true,
		},
		{
			name:     "uint64",
			input:    uint64(42),
			want:     "42",
			wantBool: true,
		},
		{
			name:     "string passthrough",
			input:    "hello",
			want:     "hello",
			wantBool: false,
		},
		{
			name:     "int passthrough",
			input:    int(42),
			want:     "42",
			wantBool: true,
		},
		{
			name:     "bool passthrough",
			input:    true,
			want:     true,
			wantBool: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, gotBool := unifyInt(tt.input)
			if got != tt.want {
				t.Errorf("unifyInt(%v) = %v (%T), want %v (%T)", tt.input, got, got, tt.want, tt.want)
			}
			if gotBool != tt.wantBool {
				t.Errorf("unifyInt(%v) bool = %v, want %v", tt.input, gotBool, tt.wantBool)
			}
		})
	}
}

func Test_getDepthByPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "root only",
			input: "$",
			want:  0,
		},
		{
			name:  "simple child",
			input: "$.items",
			want:  0,
		},
		{
			name:  "one wildcard",
			input: "$.items[*]",
			want:  1,
		},
		{
			name:  "two wildcards",
			input: "$.items[*][*]",
			want:  2,
		},
		{
			name:  "recursive descent",
			input: "$..items",
			want:  1,
		},
		{
			name:  "recursive descent and wildcard",
			input: "$..items[*]",
			want:  2,
		},
		{
			name:  "wildcard inside quotes ignored",
			input: "$.foo.'bar[*]'.items",
			want:  0,
		},
		{
			name:  "double dot inside quotes ignored",
			input: "$.foo.'bar..baz'.items",
			want:  0,
		},
		{
			name:  "quoted ignored unquoted counted",
			input: "$.foo.'bar[*]'.items[*]",
			want:  1,
		},
		{
			name:  "escaped quote inside quoted segment",
			input: "$.foo.'it\\'s[*]'.items",
			want:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := getDepthByPath(tt.input)
			if got != tt.want {
				t.Errorf("getDepthByPath(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func Test_compareKey(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		key          any
		keyNodeValue any
		want         bool
	}{
		{
			name:         "same string",
			key:          "name",
			keyNodeValue: "name",
			want:         true,
		},
		{
			name:         "different string",
			key:          "name",
			keyNodeValue: "age",
			want:         false,
		},
		{
			name:         "int64 vs int",
			key:          int64(1),
			keyNodeValue: int(1),
			want:         true,
		},
		{
			name:         "uint64 vs int",
			key:          uint64(1),
			keyNodeValue: int(1),
			want:         true,
		},
		{
			name:         "int64 vs uint64",
			key:          int64(5),
			keyNodeValue: uint64(5),
			want:         true,
		},
		{
			name:         "different int values",
			key:          int64(1),
			keyNodeValue: int64(2),
			want:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := compareKey(tt.key, tt.keyNodeValue)
			if got != tt.want {
				t.Errorf("compareKey(%v, %v) = %v, want %v", tt.key, tt.keyNodeValue, got, tt.want)
			}
		})
	}
}
