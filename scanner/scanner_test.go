package scanner_test

import (
	"iter"
	"strings"
	"testing"

	"deedles.dev/extract/scanner"
)

func checkTokens(t *testing.T, got iter.Seq2[scanner.Token, error], ex []any) {
	t.Helper()

	var i int
	for tok, err := range got {
		if err != nil {
			t.Fatal(err)
		}

		if tok.Val != ex[i] {
			t.Fatal(tok)
		}
		i++
	}
}

func TestScan(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output []any
	}{
		{"Simple", `("test" 30 'a' 1.2 push)`, []any{scanner.Lparen{}, scanner.String("test"), scanner.Int(30), scanner.Int('a'), scanner.Float(1.2), scanner.Ident("push"), scanner.Rparen{}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			checkTokens(t, scanner.Scan(strings.NewReader(test.input)), test.output)
		})
	}
}
