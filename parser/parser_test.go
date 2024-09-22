package parser_test

import (
	"iter"
	"strings"
	"testing"

	"deedles.dev/extract"
	"deedles.dev/extract/literal"
	"deedles.dev/extract/parser"
)

func checkList(t *testing.T, got literal.List, ex literal.List) {
	next, stop := iter.Pull(ex.All())
	defer stop()

	for g := range got.All() {
		e, ok := next()
		if !ok {
			t.Fatal(g)
		}

		switch g := g.(type) {
		case literal.List:
			checkList(t, g, e.(literal.List))
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
		output literal.List
	}{
		{"Simple", `(IO.println "This is a test.")`, literal.List{List: extract.ListOf(
			literal.List{List: extract.ListOf(
				literal.Ref{In: extract.MakeAtom("IO"), Name: extract.MakeIdent("println")},
				"This is a test.",
			)},
		)}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			list, err := parser.Parse(strings.NewReader(test.input))
			if err != nil {
				t.Fatal(err)
			}
			checkList(t, literal.List{List: list}, test.output)
		})
	}
}
