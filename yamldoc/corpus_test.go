package yamldoc_test

import (
	"bufio"
	"bytes"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
	"github.com/suzuki-shunsuke/go-yamledit/yamldoc"
)

// The tests in this file run against YAML files listed in the file $YAMLDOC_CORPUS.
// They are skipped if the environment variable isn't set.
//
//	find ~/repos -name '*.yaml' > /tmp/corpus.txt
//	YAMLDOC_CORPUS=/tmp/corpus.txt go test ./yamldoc -run Corpus -v

func readCorpus(t *testing.T) []string {
	t.Helper()
	list := os.Getenv("YAMLDOC_CORPUS")
	if list == "" {
		t.Skip("YAMLDOC_CORPUS isn't set")
	}
	fp, err := os.Open(list) //nolint:gosec
	if err != nil {
		t.Fatal(err)
	}
	defer fp.Close()
	var paths []string
	sc := bufio.NewScanner(fp)
	for sc.Scan() {
		paths = append(paths, sc.Text())
	}
	return paths
}

// TestRoundTripCorpus checks the output without changes is same as goccy/go-yaml's output.
func TestRoundTripCorpus(t *testing.T) {
	t.Parallel()
	total, same := 0, 0
	for _, path := range readCorpus(t) {
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		file, err := parser.ParseBytes(b, parser.ParseComments)
		if err != nil {
			continue
		}
		total++
		f, err := yamldoc.ParseAll(b)
		if err != nil {
			t.Errorf("%s: parse: %v", path, err)
			continue
		}
		got, err := f.ToString()
		if err != nil {
			t.Errorf("%s: to string: %v", path, err)
			continue
		}
		if got != file.String() {
			t.Errorf("%s: differs from goccy/go-yaml's output", path)
			continue
		}
		same++
	}
	t.Logf("total=%d same=%d", total, same)
}

func mutateNode(t *testing.T, n yamldoc.Node) {
	t.Helper()
	switch v := n.(type) {
	case *yamldoc.Map:
		for _, p := range v.Items {
			mutateNode(t, p.Value)
		}
		slices.Reverse(v.Items)
		if err := v.Set("zz_yamldoc", map[string]any{"a": []any{1, map[string]any{"b": "x\ny\n"}}}); err != nil {
			t.Fatal(err)
		}
	case *yamldoc.Seq:
		for _, item := range v.Items {
			mutateNode(t, item)
		}
		if err := v.Add(map[string]any{"c": 1, "d": []int{1}}); err != nil {
			t.Fatal(err)
		}
	}
}

func mutateValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		for k, e := range x {
			x[k] = mutateValue(e)
		}
		x["zz_yamldoc"] = map[string]any{"a": []any{uint64(1), map[string]any{"b": "x\ny\n"}}}
		return x
	case []any:
		for i, e := range x {
			x[i] = mutateValue(e)
		}
		return append(x, map[string]any{"c": uint64(1), "d": []any{uint64(1)}})
	}
	return v
}

// mutateFile reverses all maps, adds a pair to all maps, and adds an item to all sequences.
// Then it checks the output is valid YAML and has the expected data.
func mutateFile(t *testing.T, path string, b []byte) bool {
	t.Helper()
	var want any
	if err := yaml.Unmarshal(b, &want); err != nil {
		return false
	}
	doc, err := yamldoc.Parse(b)
	if err != nil {
		return false
	}
	mutateNode(t, doc.Contents)
	want = mutateValue(want)
	s, err := doc.ToString()
	if err != nil {
		t.Errorf("%s: to string: %v", path, err)
		return true
	}
	var got any
	if err := yaml.Unmarshal([]byte(s), &got); err != nil {
		t.Errorf("%s: output is invalid: %v", path, err)
		return true
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s: output differs", path)
	}
	return true
}

func TestMutateCorpus(t *testing.T) {
	t.Parallel()
	total := 0
	for _, path := range readCorpus(t) {
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		// Anchors, aliases, and merge keys are skipped because edits of an anchor affect aliases.
		// Multiple documents are skipped for simplicity.
		if bytes.Contains(b, []byte("&")) || bytes.Contains(b, []byte("<<")) || bytes.Contains(b, []byte("\n---")) {
			continue
		}
		if mutateFile(t, path, b) {
			total++
		}
	}
	t.Logf("total=%d", total)
}
