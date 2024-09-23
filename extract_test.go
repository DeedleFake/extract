package extract_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"deedles.dev/extract"
	"deedles.dev/extract/parser"
)

func runScript(t *testing.T, src string, checkErrors bool) any {
	s, err := parser.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}

	r := extract.New(context.Background())
	_, result := extract.Run(r, s.All())
	if err, ok := result.(error); ok && checkErrors {
		t.Fatal(err)
	}

	return result
}

func TestSimpleScript(t *testing.T) {
	const src = `"This is a test."`
	result := runScript(t, src, true)
	if result != "This is a test." {
		t.Fatalf("%#v", result)
	}
}

func TestSingleCall(t *testing.T) {
	const src = `(String.to_upper "test")`
	result := runScript(t, src, true)
	if result != "TEST" {
		t.Fatalf("%#v", result)
	}
}

func TestStringFormat(t *testing.T) {
	const src = `(String.format "This is a %v." "test")`
	result := runScript(t, src, true)
	if result != "This is a test." {
		t.Fatalf("%#v", result)
	}
}

func TestDefModule(t *testing.T) {
	const src = `
	(defmodule Test
		(def (inc v) (add v 1))
	)

	(Test.inc 2)
	`
	result := runScript(t, src, true)
	if result != int64(3) {
		t.Fatalf("%#v", result)
	}
}

func BenchmarkDefModule(b *testing.B) {
	for range b.N {
		const src = `
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

func TestIndirectFunctionCall(t *testing.T) {
	const src = `
	(defmodule Test
		(def (get _) (func (plus a b) (add a b)))
	)

	((Test.get ()) 1 2)
	`
	result := runScript(t, src, true)
	if result != int64(3) {
		t.Fatalf("%#v", result)
	}
}

func TestErrPatternMatch(t *testing.T) {
	const src = `
	(defmodule Test
		(def (test 1) ())
	)

	(Test.test 2)
	`
	result := runScript(t, src, false)
	if err, ok := result.(error); !ok || !errors.Is(err, extract.ErrPatternMatch) {
		t.Fatalf("%#v", result)
	}
}
