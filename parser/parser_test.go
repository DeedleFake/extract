package parser_test

import (
	"iter"
	"strings"
	"testing"

	"deedles.dev/extract"
	"deedles.dev/extract/parser"
)

func checkList(t *testing.T, got *extract.List, ex *extract.List) {
	next, stop := iter.Pull(ex.All())
	defer stop()

	for g := range got.All() {
		e, ok := next()
		if !ok {
			t.Fatal(g)
		}

		switch g := g.(type) {
		case *extract.List:
			checkList(t, g, e.(*extract.List))
		default:
			if g != e {
				t.Fatalf("%#v != %#v", g, e)
			}
		}
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output *extract.List
	}{
		{"Simple", `(IO.println "This is a test.")`, extract.ListOf(
			extract.ListOf(
				extract.ModuleIdent{Module: extract.Atom("IO"), Ident: extract.Ident("println")},
				extract.String("This is a test."),
			),
		)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			list, err := parser.Parse(strings.NewReader(test.input))
			if err != nil {
				t.Fatal(err)
			}
			checkList(t, list, test.output)
		})
	}
}
