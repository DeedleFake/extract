package extract_test

import (
	"strings"
	"testing"

	"deedles.dev/extract"
	"deedles.dev/extract/parser"
)

func TestSimpleScript(t *testing.T) {
	src := `"This is a test."`
	s, err := parser.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	r := extract.NewRuntime()
	result, err := s.Run(r.Context())
	if err != nil {
		t.Fatal(err)
	}
	if result != "This is a test." {
		t.Fatalf("%q", result)
	}
}
