package scanner_test

import (
	"strings"
	"testing"

	"deedles.dev/extract/scanner"
)

func checkTokens(t *testing.T, s *scanner.Scanner, ex []any) {
	var i int
	for tok := range s.All() {
		if tok.Val != ex[i] {
			t.Fatal(tok)
		}
		i++
	}
	if s.Err() != nil {
		t.Fatal(s.Err())
	}
	if i != len(ex) {
		t.Fatal(i)
	}
}

func TestScan(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output []any
	}{
		{"Simple", `("test" 30 'a' 1.2 :test2 Test3.push)`, []any{
			scanner.Lparen{},
			scanner.String("test"),
			scanner.Int(30),
			scanner.Int('a'),
			scanner.Float(1.2),
			scanner.Atom("test2"),
			scanner.Atom("Test3"),
			scanner.Dot{},
			scanner.Ident("push"),
			scanner.Rparen{},
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			checkTokens(t, scanner.New(strings.NewReader(test.input)), test.output)
		})
	}
}
