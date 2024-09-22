// Package parser implements a parser for Extract code.
package parser

import (
	"errors"
	"fmt"
	"io"

	"deedles.dev/extract"
	"deedles.dev/extract/literal"
	"deedles.dev/extract/scanner"
)

// Parse parses an Extract script from r.
func Parse(r io.Reader) (*extract.List, error) {
	return ParseScanner(scanner.New(r))
}

// ParseScanner parses an Extract script from s.
func ParseScanner(s *scanner.Scanner) (*extract.List, error) {
	p := parser{s: s}
	return p.Parse()
}

type parser struct {
	s   *scanner.Scanner
	tok scanner.Token
}

func (p *parser) Parse() (list *extract.List, err error) {
	defer func() {
		switch r := recover().(type) {
		case nil:
		case raise:
			err = r.err
		default:
			panic(r)
		}
	}()

	return p.listInner(), nil
}

type raise struct{ err error }

func (p *parser) raise(err error) {
	panic(raise{err: err})
}

func (p *parser) raiseUnexpectedEOF() {
	p.raise(io.ErrUnexpectedEOF)
}

func (p *parser) raiseUnexpectedToken(got scanner.Token, ex any) {
	p.raise(&UnexpectedTokenError{
		Line:     got.Line,
		Col:      got.Col,
		Got:      got.Val,
		Expected: ex,
	})
}

func (p *parser) scan() scanner.Token {
	if p.tok.Val != nil {
		tok := p.tok
		p.tok.Val = nil
		return tok
	}

	if !p.s.Scan() {
		p.raiseUnexpectedEOF()
		return scanner.Token{}
	}

	return p.s.Token()
}

func (p *parser) unscan(tok scanner.Token) {
	if p.tok.Val != nil {
		panic("unscanned twice")
	}

	p.tok = tok
}

func (p *parser) peek() any {
	if p.tok.Val != nil {
		return p.tok.Val
	}

	if !p.s.Scan() {
		return nil
	}

	p.tok = p.s.Token()
	return p.tok.Val
}

func expect[T any](p *parser) (tok scanner.Token, v T) {
	got := p.scan()
	if v, ok := got.Val.(T); ok {
		return got, v
	}

	p.raiseUnexpectedToken(got, nil)
	return tok, v
}

func (p *parser) list() literal.List {
	expect[scanner.Lparen](p)
	list := p.listInner()
	expect[scanner.Rparen](p)
	return literal.List{List: list}
}

func (p *parser) listInner() *extract.List {
	var exprs []any
	for p.peek() != (scanner.Rparen{}) && p.peek() != nil {
		exprs = append(exprs, p.expr())
	}
	return extract.ListOf(exprs...)
}

func (p *parser) expr() (expr any) {
	tok := p.scan()
	switch t := tok.Val.(type) {
	case scanner.Int:
		expr = literal.Int(t)
	case scanner.Float:
		expr = literal.Float(t)
	case scanner.String:
		expr = literal.String(t)
	case scanner.Atom:
		expr = extract.MakeAtom(string(t))
	case scanner.Ident:
		expr = extract.MakeIdent(string(t))
	case scanner.Lparen:
		p.unscan(tok)
		expr = p.list()
	default:
		p.raiseUnexpectedToken(p.scan(), nil)
		return nil
	}

	if p.peek() == (scanner.Dot{}) {
		expr = p.ref(expr)
	}

	return expr
}

func (p *parser) ref(in any) literal.Ref {
	expect[scanner.Dot](p)
	switch name := p.expr().(type) {
	case extract.Ident:
		return literal.Ref{In: in, Name: name}
	default:
		p.raise(errors.New("last element of a ref must be an identifier"))
		return literal.Ref{}
	}
}

// UnexpectedTokenError is returned from an attempt to parse a script
// if the script has a token somewhere that it shouldn't be. If there
// was a specific token that was supposed to be there, it will be
// indicated with the Expected field.
type UnexpectedTokenError struct {
	Line, Col int
	Got       any
	Expected  any
}

func (err *UnexpectedTokenError) Error() string {
	if err.Expected == nil {
		return fmt.Sprintf("unexpected token %q (%[1]T) at %v:%v", err.Got, err.Line, err.Col)
	}
	return fmt.Sprintf("unexpected token %q (%[1]T) at %v:%v, expected %q (%[4]T)", err.Got, err.Line, err.Col, err.Expected)
}
