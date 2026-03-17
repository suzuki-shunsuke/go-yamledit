package mag

import "testing"

func Test_unifyInt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input any
		want  any
	}{
		{
			name:  "int64",
			input: int64(42),
			want:  int(42),
		},
		{
			name:  "uint64",
			input: uint64(42),
			want:  int(42),
		},
		{
			name:  "string passthrough",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "int passthrough",
			input: int(42),
			want:  int(42),
		},
		{
			name:  "bool passthrough",
			input: true,
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := unifyInt(tt.input)
			if got != tt.want {
				t.Errorf("unifyInt(%v) = %v (%T), want %v (%T)", tt.input, got, got, tt.want, tt.want)
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
