package extract_test

import (
	"context"
	"strings"
	"testing"

	"deedles.dev/extract"
	"deedles.dev/extract/parser"
)

func runScript(t *testing.T, src string) any {
	s, err := parser.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}

	r := extract.New(context.Background())
	_, result := extract.Run(r, s.All())
	if err, ok := result.(error); ok {
		t.Fatal(err)
	}

	return result
}

func TestSimpleScript(t *testing.T) {
	src := `"This is a test."`
	result := runScript(t, src)
	if result != "This is a test." {
		t.Fatalf("%#v", result)
	}
}

func TestSingleCall(t *testing.T) {
	src := `(String.to_upper "test")`
	result := runScript(t, src)
	if result != "TEST" {
		t.Fatalf("%#v", result)
	}
}

func TestStringFormat(t *testing.T) {
	src := `(String.format "This is a %v." "test")`
	result := runScript(t, src)
	if result != "This is a test." {
		t.Fatalf("%#v", result)
	}
}

func TestDefModule(t *testing.T) {
	src := `
	(defmodule Test
		(def (inc v) (add v 1))
	)

	(Test.inc 2)
	`
	result := runScript(t, src)
	if result != int64(3) {
		t.Fatalf("%#v", result)
	}
}

func BenchmarkDefModule(b *testing.B) {
	for range b.N {
		src := `
		(defmodule Test
			(def (inc v) (add v 1))
		)

		(Test.inc 2)
		`
		s, _ := parser.Parse(strings.NewReader(src))
		r := extract.New(context.Background())
		extract.Run(r, s.All())
	}
}
